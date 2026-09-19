package auth_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/snowaa-desigram/backend/services/go/internal/auth/store"

	"github.com/alicebob/miniredis/v2"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

// Контрактные тесты хранилищ: одна таблица проверок гоняется по каждой реализации.
// memory и Redis (miniredis) — всегда; GORM/MySQL — при AUTH_TEST_MYSQL_DSN (CI поднимает mysql service).

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("AUTH_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("AUTH_TEST_MYSQL_DSN is not set")
	}
	db, err := store.OpenDB(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	// чистим за собой (FK: сначала токены)
	t.Cleanup(func() {
		db.Exec("DELETE FROM auth_refresh_tokens")
		db.Exec("DELETE FROM auth_users")
	})
	return db
}

func testRedis(t *testing.T) *redis.Redis {
	t.Helper()
	mr := miniredis.RunT(t)
	return redis.New(mr.Addr())
}

// ---- store.UserStore ----

func TestUserStore(t *testing.T) {
	t.Run("memory", func(t *testing.T) { runUserStoreTests(t, store.NewMemoryUserStore()) })
	t.Run("gorm", func(t *testing.T) { runUserStoreTests(t, store.NewGormUserStore(testDB(t), testRedis(t))) })
}

func runUserStoreTests(t *testing.T, s store.UserStore) {
	ctx := context.Background()
	now := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	u := &store.User{ID: "11111111-1111-1111-1111-111111111111", Email: "a@example.com", PasswordHash: "h1", CreatedAt: now, UpdatedAt: now}

	// промах по email до создания — кеш «нет» не должен пережить Create
	if _, err := s.FindByEmail(ctx, u.Email); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("FindByEmail before create: %v", err)
	}
	if err := s.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}
	dup := &store.User{ID: "22222222-2222-2222-2222-222222222222", Email: u.Email, PasswordHash: "h2", CreatedAt: now, UpdatedAt: now}
	if err := s.Create(ctx, dup); !errors.Is(err, store.ErrDuplicate) {
		t.Fatalf("Create duplicate: %v, want store.ErrDuplicate", err)
	}

	got, err := s.FindByEmail(ctx, u.Email)
	if err != nil || got.ID != u.ID || got.PasswordHash != "h1" || got.Verified() {
		t.Fatalf("FindByEmail: %+v, %v", got, err)
	}
	if got, err := s.FindByID(ctx, u.ID); err != nil || got.Email != u.Email {
		t.Fatalf("FindByID: %+v, %v", got, err)
	}
	if _, err := s.FindByID(ctx, "missing"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("FindByID missing: %v", err)
	}

	verified := now.Add(time.Minute)
	got.PasswordHash, got.EmailVerifiedAt, got.UpdatedAt = "h3", &verified, verified
	if err := s.Update(ctx, got); err != nil {
		t.Fatalf("Update: %v", err)
	}
	// после Update и по id, и по email видим новые данные (кеш сброшен)
	for _, find := range []func() (*store.User, error){
		func() (*store.User, error) { return s.FindByID(ctx, u.ID) },
		func() (*store.User, error) { return s.FindByEmail(ctx, u.Email) },
	} {
		got, err := find()
		if err != nil || got.PasswordHash != "h3" || !got.Verified() || !got.EmailVerifiedAt.Equal(verified) {
			t.Fatalf("after Update: %+v, %v", got, err)
		}
	}
}

// ---- store.RefreshTokenStore ----

func TestRefreshTokenStore(t *testing.T) {
	t.Run("memory", func(t *testing.T) { runRefreshTokenStoreTests(t, store.NewMemoryRefreshTokenStore(), nil) })
	t.Run("gorm", func(t *testing.T) {
		db := testDB(t)
		runRefreshTokenStoreTests(t, store.NewGormRefreshTokenStore(db), store.NewGormUserStore(db, testRedis(t)))
	})
}

func runRefreshTokenStoreTests(t *testing.T, s store.RefreshTokenStore, users store.UserStore) {
	ctx := context.Background()
	now := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	const userA, userB = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	if users != nil { // FK в MySQL
		for _, id := range []string{userA, userB} {
			if err := users.Create(ctx, &store.User{ID: id, Email: id + "@example.com", PasswordHash: "h", CreatedAt: now, UpdatedAt: now}); err != nil {
				t.Fatal(err)
			}
		}
	}
	mk := func(id, user, hash string) *store.RefreshToken {
		return &store.RefreshToken{ID: id, UserID: user, TokenHash: hash, ExpiresAt: now.Add(time.Hour), CreatedAt: now}
	}
	for _, tok := range []*store.RefreshToken{mk("t1", userA, "h1"), mk("t2", userA, "h2"), mk("t3", userB, "h3")} {
		if err := s.Create(ctx, tok); err != nil {
			t.Fatalf("Create %s: %v", tok.ID, err)
		}
	}

	got, err := s.FindByHash(ctx, "h1")
	if err != nil || got.ID != "t1" || got.UserID != userA || got.RevokedAt != nil || !got.ExpiresAt.Equal(now.Add(time.Hour)) {
		t.Fatalf("FindByHash: %+v, %v", got, err)
	}
	if _, err := s.FindByHash(ctx, "nope"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("FindByHash missing: %v", err)
	}

	at := now.Add(time.Minute)
	if err := s.Revoke(ctx, "t1", at); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.FindByHash(ctx, "h1"); got.RevokedAt == nil || !got.RevokedAt.Equal(at) {
		t.Fatalf("after Revoke: %+v", got)
	}
	// повторный Revoke не двигает время
	if err := s.Revoke(ctx, "t1", at.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.FindByHash(ctx, "h1"); !got.RevokedAt.Equal(at) {
		t.Fatalf("second Revoke changed time: %+v", got)
	}

	if err := s.Delete(ctx, "t2"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.FindByHash(ctx, "h2"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("after Delete: %v", err)
	}
	if err := s.Delete(ctx, "t2"); err != nil { // идемпотентно
		t.Fatal(err)
	}

	if err := s.DeleteAllForUser(ctx, userA); err != nil {
		t.Fatal(err)
	}
	if _, err := s.FindByHash(ctx, "h1"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("after DeleteAllForUser: %v", err)
	}
	if _, err := s.FindByHash(ctx, "h3"); err != nil {
		t.Fatalf("other user's token must survive: %v", err)
	}
}

// ---- store.CodeStore ----

func TestCodeStore(t *testing.T) {
	t.Run("memory", func(t *testing.T) {
		now := time.Now()
		runCodeStoreTests(t, store.NewMemoryCodeStore(func() time.Time { return now }), func(d time.Duration) { now = now.Add(d) })
	})
	t.Run("redis", func(t *testing.T) {
		mr := miniredis.RunT(t)
		runCodeStoreTests(t, store.NewRedisCodeStore(redis.New(mr.Addr())), mr.FastForward)
	})
}

func runCodeStoreTests(t *testing.T, s store.CodeStore, advance func(time.Duration)) {
	ctx := context.Background()
	const email = "c@example.com"

	if _, _, err := s.Get(ctx, store.PurposeRegister, email); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("Get empty: %v", err)
	}
	if _, err := s.IncrAttempts(ctx, store.PurposeRegister, email); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("IncrAttempts without code: %v", err)
	}

	if err := s.Put(ctx, store.PurposeRegister, email, "hash1", 10*time.Minute); err != nil {
		t.Fatal(err)
	}
	hash, attempts, err := s.Get(ctx, store.PurposeRegister, email)
	if err != nil || hash != "hash1" || attempts != 0 {
		t.Fatalf("Get: %q %d %v", hash, attempts, err)
	}
	// другое назначение — отдельный код
	if _, _, err := s.Get(ctx, store.PurposePasswordReset, email); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("Get other purpose: %v", err)
	}

	for want := 1; want <= 2; want++ {
		if n, err := s.IncrAttempts(ctx, store.PurposeRegister, email); err != nil || n != want {
			t.Fatalf("IncrAttempts: %d %v, want %d", n, err, want)
		}
	}
	if _, attempts, _ := s.Get(ctx, store.PurposeRegister, email); attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
	// новый код обнуляет попытки
	if err := s.Put(ctx, store.PurposeRegister, email, "hash2", 10*time.Minute); err != nil {
		t.Fatal(err)
	}
	if hash, attempts, _ := s.Get(ctx, store.PurposeRegister, email); hash != "hash2" || attempts != 0 {
		t.Fatalf("after re-Put: %q %d", hash, attempts)
	}
	if _, err := s.IncrAttempts(ctx, store.PurposeRegister, email); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(ctx, store.PurposeRegister, email); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.Get(ctx, store.PurposeRegister, email); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("after Delete: %v", err)
	}
	// попытки удалены вместе с кодом
	if err := s.Put(ctx, store.PurposeRegister, email, "hash3", time.Minute); err != nil {
		t.Fatal(err)
	}
	if _, attempts, _ := s.Get(ctx, store.PurposeRegister, email); attempts != 0 {
		t.Fatalf("attempts survived Delete: %d", attempts)
	}
	advance(time.Minute + time.Second)
	if _, _, err := s.Get(ctx, store.PurposeRegister, email); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("after TTL: %v", err)
	}

	// cooldown
	if ok, err := s.SetCooldown(ctx, store.PurposeRegister, email, time.Minute); err != nil || !ok {
		t.Fatalf("SetCooldown first: %v %v", ok, err)
	}
	if ok, _ := s.SetCooldown(ctx, store.PurposeRegister, email, time.Minute); ok {
		t.Fatal("SetCooldown second must be false")
	}
	if ok, _ := s.SetCooldown(ctx, store.PurposePasswordReset, email, time.Minute); !ok {
		t.Fatal("cooldown is per purpose")
	}
	advance(time.Minute + time.Second)
	if ok, _ := s.SetCooldown(ctx, store.PurposeRegister, email, time.Minute); !ok {
		t.Fatal("SetCooldown after TTL must be true")
	}

	// неудачные логины
	if n, err := s.Failures(ctx, email); err != nil || n != 0 {
		t.Fatalf("Failures empty: %d %v", n, err)
	}
	for want := 1; want <= 3; want++ {
		if n, err := s.IncrFailures(ctx, email, time.Minute); err != nil || n != want {
			t.Fatalf("IncrFailures: %d %v, want %d", n, err, want)
		}
	}
	if n, _ := s.Failures(ctx, email); n != 3 {
		t.Fatalf("Failures = %d", n)
	}
	if err := s.ResetFailures(ctx, email); err != nil {
		t.Fatal(err)
	}
	if n, _ := s.Failures(ctx, email); n != 0 {
		t.Fatalf("Failures after reset = %d", n)
	}
	// окно не продлевается повторными неудачами
	for _, d := range []time.Duration{30 * time.Second, 31 * time.Second} {
		if _, err := s.IncrFailures(ctx, email, time.Minute); err != nil {
			t.Fatal(err)
		}
		advance(d)
	}
	if n, _ := s.Failures(ctx, email); n != 0 {
		t.Fatalf("Failures after window = %d, want 0", n)
	}
}
