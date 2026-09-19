package store

import (
	"context"
	"errors"
	"time"
)

// Модели и интерфейсы хранилищ: реализации — gorm.go, redis.go, memory.go (тесты/dev).

var (
	ErrNotFound  = errors.New("not found")
	ErrDuplicate = errors.New("duplicate")
)

// User — и доменная структура, и GORM-модель (таблица auth_users).
type User struct {
	ID              string     `gorm:"type:char(36);primaryKey"`
	Email           string     `gorm:"type:varchar(255);not null;uniqueIndex:uq_auth_users_email"`
	PasswordHash    string     `gorm:"type:varchar(255);not null"`
	EmailVerifiedAt *time.Time `gorm:"type:datetime"`
	CreatedAt       time.Time  `gorm:"type:datetime;not null"`
	UpdatedAt       time.Time  `gorm:"type:datetime;not null"`
}

func (User) TableName() string { return "auth_users" }

func (u *User) Verified() bool { return u.EmailVerifiedAt != nil }

// RefreshToken — GORM-модель (таблица auth_refresh_tokens). Хранится только sha256 токена.
type RefreshToken struct {
	ID        string     `gorm:"type:char(36);primaryKey"`
	UserID    string     `gorm:"type:char(36);not null;index:idx_auth_refresh_tokens_user"`
	TokenHash string     `gorm:"type:char(64);not null;uniqueIndex:uq_auth_refresh_tokens_hash"`
	ExpiresAt time.Time  `gorm:"type:datetime;not null"`
	RevokedAt *time.Time `gorm:"type:datetime"`
	CreatedAt time.Time  `gorm:"type:datetime;not null"`

	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (RefreshToken) TableName() string { return "auth_refresh_tokens" }

type UserStore interface {
	// Create возвращает ErrDuplicate, если email уже занят.
	Create(ctx context.Context, u *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	// Update пишет PasswordHash, EmailVerifiedAt, UpdatedAt.
	Update(ctx context.Context, u *User) error
}

// RefreshTokenStore: Revoke — при ротации (запись остаётся: повторное использование = признак кражи),
// Delete/DeleteAllForUser — logout и сброс пароля (запись исчезает, каскада на другие сессии нет).
type RefreshTokenStore interface {
	Create(ctx context.Context, t *RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*RefreshToken, error)
	Revoke(ctx context.Context, id string, at time.Time) error
	Delete(ctx context.Context, id string) error
	DeleteAllForUser(ctx context.Context, userID string) error
}

// CodePurpose — зачем выслан код; коды разных назначений не взаимозаменяемы.
type CodePurpose string

const (
	PurposeRegister      CodePurpose = "register"
	PurposePasswordReset CodePurpose = "password_reset"
)

// CodeStore — коды подтверждения и счётчики (Redis с TTL).
type CodeStore interface {
	// Put сохраняет хеш кода на ttl и обнуляет счётчик попыток.
	Put(ctx context.Context, purpose CodePurpose, email, hash string, ttl time.Duration) error
	// Get возвращает хеш и число неудачных попыток; ErrNotFound — кода нет или истёк.
	Get(ctx context.Context, purpose CodePurpose, email string) (hash string, attempts int, err error)
	// IncrAttempts увеличивает счётчик неудачных попыток и возвращает новое значение.
	IncrAttempts(ctx context.Context, purpose CodePurpose, email string) (int, error)
	Delete(ctx context.Context, purpose CodePurpose, email string) error
	// SetCooldown ставит флаг «код недавно отправлен»; false — флаг уже стоит.
	SetCooldown(ctx context.Context, purpose CodePurpose, email string, ttl time.Duration) (bool, error)

	// Неудачные логины по email в скользящем окне window.
	Failures(ctx context.Context, email string) (int, error)
	IncrFailures(ctx context.Context, email string, window time.Duration) (int, error)
	ResetFailures(ctx context.Context, email string) error
}
