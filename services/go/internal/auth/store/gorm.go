package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/syncx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MySQL через GORM. Пользователи — cache-aside в Redis (go-zero cache): в SQL только промахи и записи,
// как в core. Refresh-токены — без кеша: одноразовые, читаются один раз.

const userCacheTTL = 24 * time.Hour

// OpenDB открывает GORM-соединение; DSN — go-sql-driver (user:pass@tcp(host:3306)/db?parseTime=true).
func OpenDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		TranslateError: true, // ErrDuplicatedKey вместо кода 1062
		Logger:         logger.New(gormLogWriter{}, logger.Config{SlowThreshold: 200 * time.Millisecond, LogLevel: logger.Warn}),
	})
	if err != nil {
		return nil, fmt.Errorf("gorm open: %w", err)
	}
	return db, nil
}

// Migrate — GORM AutoMigrate. Запускается одноразовым compose-сервисом auth-migrate до старта реплик.
func Migrate(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).AutoMigrate(&User{}, &RefreshToken{})
}

type gormLogWriter struct{}

func (gormLogWriter) Printf(format string, args ...any) { logx.Infof(format, args...) }

// ---- users ----

type GormUserStore struct {
	db    *gorm.DB
	cache cache.Cache
}

func NewGormUserStore(db *gorm.DB, rds *redis.Redis) *GormUserStore {
	c := cache.NewNode(rds, syncx.NewSingleFlight(), cache.NewStat("auth-users"), ErrNotFound, cache.WithExpiry(userCacheTTL))
	return &GormUserStore{db: db, cache: c}
}

func userIDKey(id string) string       { return "auth:user:id:" + id }
func userEmailKey(email string) string { return "auth:user:email:" + email }

func (s *GormUserStore) Create(ctx context.Context, u *User) error {
	if err := s.db.WithContext(ctx).Create(u).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrDuplicate
		}
		return err
	}
	// Промах по email мог закешироваться как «нет» — сбрасываем.
	return s.cache.DelCtx(ctx, userIDKey(u.ID), userEmailKey(u.Email))
}

func (s *GormUserStore) FindByID(ctx context.Context, id string) (*User, error) {
	var u User
	err := s.cache.TakeCtx(ctx, &u, userIDKey(id), func(v any) error {
		return notFound(s.db.WithContext(ctx).Take(v, "id = ?", id).Error)
	})
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByEmail: email → id (кеш) → пользователь (кеш), чтобы у пользователя была одна копия в кеше.
func (s *GormUserStore) FindByEmail(ctx context.Context, email string) (*User, error) {
	var id string
	err := s.cache.TakeCtx(ctx, &id, userEmailKey(email), func(v any) error {
		var ids []string
		if err := s.db.WithContext(ctx).Model(&User{}).Where("email = ?", email).Limit(1).Pluck("id", &ids).Error; err != nil {
			return err
		}
		if len(ids) == 0 {
			return ErrNotFound
		}
		*v.(*string) = ids[0]
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.FindByID(ctx, id)
}

func (s *GormUserStore) Update(ctx context.Context, u *User) error {
	err := s.db.WithContext(ctx).Model(u).Select("password_hash", "email_verified_at", "updated_at").Updates(u).Error
	if err != nil {
		return err
	}
	return s.cache.DelCtx(ctx, userIDKey(u.ID), userEmailKey(u.Email))
}

// ---- refresh tokens ----

type GormRefreshTokenStore struct {
	db *gorm.DB
}

func NewGormRefreshTokenStore(db *gorm.DB) *GormRefreshTokenStore {
	return &GormRefreshTokenStore{db: db}
}

func (s *GormRefreshTokenStore) Create(ctx context.Context, t *RefreshToken) error {
	return s.db.WithContext(ctx).Omit("User").Create(t).Error
}

func (s *GormRefreshTokenStore) FindByHash(ctx context.Context, hash string) (*RefreshToken, error) {
	var t RefreshToken
	if err := notFound(s.db.WithContext(ctx).Take(&t, "token_hash = ?", hash).Error); err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *GormRefreshTokenStore) Revoke(ctx context.Context, id string, at time.Time) error {
	return s.db.WithContext(ctx).Model(&RefreshToken{}).
		Where("id = ? AND revoked_at IS NULL", id).Update("revoked_at", at).Error
}

func (s *GormRefreshTokenStore) Delete(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&RefreshToken{}, "id = ?", id).Error
}

func (s *GormRefreshTokenStore) DeleteAllForUser(ctx context.Context, userID string) error {
	return s.db.WithContext(ctx).Delete(&RefreshToken{}, "user_id = ?", userID).Error
}

func notFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}
