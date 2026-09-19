package auth

import (
	"context"
	"sync"
	"time"
)

// In-memory реализации хранилищ для тестов. now — инжектируемые часы (TTL считается по ним).

type MemoryUserStore struct {
	mu    sync.Mutex
	users map[string]*User // по id
}

func NewMemoryUserStore() *MemoryUserStore {
	return &MemoryUserStore{users: map[string]*User{}}
}

func (s *MemoryUserStore) Create(_ context.Context, u *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, e := range s.users {
		if e.Email == u.Email {
			return ErrDuplicate
		}
	}
	c := *u
	s.users[u.ID] = &c
	return nil
}

func (s *MemoryUserStore) FindByID(_ context.Context, id string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if u, ok := s.users[id]; ok {
		c := *u
		return &c, nil
	}
	return nil, ErrNotFound
}

func (s *MemoryUserStore) FindByEmail(_ context.Context, email string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range s.users {
		if u.Email == email {
			c := *u
			return &c, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryUserStore) Update(_ context.Context, u *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.users[u.ID]
	if !ok {
		return ErrNotFound
	}
	e.PasswordHash = u.PasswordHash
	e.EmailVerifiedAt = u.EmailVerifiedAt
	e.UpdatedAt = u.UpdatedAt
	return nil
}

type MemoryRefreshTokenStore struct {
	mu     sync.Mutex
	tokens map[string]*RefreshToken // по id
}

func NewMemoryRefreshTokenStore() *MemoryRefreshTokenStore {
	return &MemoryRefreshTokenStore{tokens: map[string]*RefreshToken{}}
}

func (s *MemoryRefreshTokenStore) Create(_ context.Context, t *RefreshToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	c := *t
	s.tokens[t.ID] = &c
	return nil
}

func (s *MemoryRefreshTokenStore) FindByHash(_ context.Context, hash string) (*RefreshToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, t := range s.tokens {
		if t.TokenHash == hash {
			c := *t
			return &c, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryRefreshTokenStore) Revoke(_ context.Context, id string, at time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tokens[id]
	if !ok {
		return ErrNotFound
	}
	if t.RevokedAt == nil {
		t.RevokedAt = &at
	}
	return nil
}

func (s *MemoryRefreshTokenStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, id)
	return nil
}

func (s *MemoryRefreshTokenStore) DeleteAllForUser(_ context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, t := range s.tokens {
		if t.UserID == userID {
			delete(s.tokens, id)
		}
	}
	return nil
}

type memoryEntry struct {
	value   string
	counter int
	expires time.Time
}

type MemoryCodeStore struct {
	mu   sync.Mutex
	now  func() time.Time
	data map[string]*memoryEntry
}

func NewMemoryCodeStore(now func() time.Time) *MemoryCodeStore {
	return &MemoryCodeStore{now: now, data: map[string]*memoryEntry{}}
}

func (s *MemoryCodeStore) live(key string) *memoryEntry {
	e, ok := s.data[key]
	if !ok {
		return nil
	}
	if !e.expires.After(s.now()) {
		delete(s.data, key)
		return nil
	}
	return e
}

func codeKey(p CodePurpose, email string) string { return "code:" + string(p) + ":" + email }
func cooldownKey(p CodePurpose, email string) string {
	return "cooldown:" + string(p) + ":" + email
}
func failuresKey(email string) string { return "failures:" + email }

func (s *MemoryCodeStore) Put(_ context.Context, p CodePurpose, email, hash string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[codeKey(p, email)] = &memoryEntry{value: hash, expires: s.now().Add(ttl)}
	return nil
}

func (s *MemoryCodeStore) Get(_ context.Context, p CodePurpose, email string) (string, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := s.live(codeKey(p, email))
	if e == nil {
		return "", 0, ErrNotFound
	}
	return e.value, e.counter, nil
}

func (s *MemoryCodeStore) IncrAttempts(_ context.Context, p CodePurpose, email string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := s.live(codeKey(p, email))
	if e == nil {
		return 0, ErrNotFound
	}
	e.counter++
	return e.counter, nil
}

func (s *MemoryCodeStore) Delete(_ context.Context, p CodePurpose, email string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, codeKey(p, email))
	return nil
}

func (s *MemoryCodeStore) SetCooldown(_ context.Context, p CodePurpose, email string, ttl time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.live(cooldownKey(p, email)) != nil {
		return false, nil
	}
	s.data[cooldownKey(p, email)] = &memoryEntry{expires: s.now().Add(ttl)}
	return true, nil
}

func (s *MemoryCodeStore) Failures(_ context.Context, email string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e := s.live(failuresKey(email)); e != nil {
		return e.counter, nil
	}
	return 0, nil
}

func (s *MemoryCodeStore) IncrFailures(_ context.Context, email string, window time.Duration) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := s.live(failuresKey(email))
	if e == nil {
		e = &memoryEntry{expires: s.now().Add(window)}
		s.data[failuresKey(email)] = e
	}
	e.counter++
	return e.counter, nil
}

func (s *MemoryCodeStore) ResetFailures(_ context.Context, email string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, failuresKey(email))
	return nil
}
