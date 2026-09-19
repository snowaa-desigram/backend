package store

import (
	"context"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// RedisCodeStore — коды подтверждения, cooldown и счётчики неудачных логинов. Всё с TTL.
type RedisCodeStore struct {
	rds *redis.Redis
}

func NewRedisCodeStore(rds *redis.Redis) *RedisCodeStore {
	return &RedisCodeStore{rds: rds}
}

func (s *RedisCodeStore) key(parts ...string) string {
	k := "auth"
	for _, p := range parts {
		k += ":" + p
	}
	return k
}

func seconds(d time.Duration) int {
	if s := int(d / time.Second); s > 0 {
		return s
	}
	return 1
}

func (s *RedisCodeStore) Put(ctx context.Context, p CodePurpose, email, hash string, ttl time.Duration) error {
	if _, err := s.rds.DelCtx(ctx, s.key("code", "attempts", string(p), email)); err != nil {
		return err
	}
	return s.rds.SetexCtx(ctx, s.key("code", string(p), email), hash, seconds(ttl))
}

func (s *RedisCodeStore) Get(ctx context.Context, p CodePurpose, email string) (string, int, error) {
	hash, err := s.rds.GetCtx(ctx, s.key("code", string(p), email))
	if err != nil {
		return "", 0, err
	}
	if hash == "" {
		return "", 0, ErrNotFound
	}
	attempts, err := s.getInt(ctx, s.key("code", "attempts", string(p), email))
	return hash, attempts, err
}

func (s *RedisCodeStore) IncrAttempts(ctx context.Context, p CodePurpose, email string) (int, error) {
	codeKey := s.key("code", string(p), email)
	ttl, err := s.rds.TtlCtx(ctx, codeKey)
	if err != nil {
		return 0, err
	}
	if ttl <= 0 {
		return 0, ErrNotFound
	}
	return s.incrWithTTL(ctx, s.key("code", "attempts", string(p), email), ttl)
}

func (s *RedisCodeStore) Delete(ctx context.Context, p CodePurpose, email string) error {
	_, err := s.rds.DelCtx(ctx, s.key("code", string(p), email), s.key("code", "attempts", string(p), email))
	return err
}

func (s *RedisCodeStore) SetCooldown(ctx context.Context, p CodePurpose, email string, ttl time.Duration) (bool, error) {
	return s.rds.SetnxExCtx(ctx, s.key("cooldown", string(p), email), "1", seconds(ttl))
}

func (s *RedisCodeStore) Failures(ctx context.Context, email string) (int, error) {
	return s.getInt(ctx, s.key("login", "fail", email))
}

func (s *RedisCodeStore) IncrFailures(ctx context.Context, email string, window time.Duration) (int, error) {
	return s.incrWithTTL(ctx, s.key("login", "fail", email), seconds(window))
}

func (s *RedisCodeStore) ResetFailures(ctx context.Context, email string) error {
	_, err := s.rds.DelCtx(ctx, s.key("login", "fail", email))
	return err
}

func (s *RedisCodeStore) getInt(ctx context.Context, key string) (int, error) {
	v, err := s.rds.GetCtx(ctx, key)
	if err != nil || v == "" {
		return 0, err
	}
	return strconv.Atoi(v)
}

// incrWithTTL — INCR; TTL ставится только при создании ключа (первое значение), чтобы окно не продлевалось.
func (s *RedisCodeStore) incrWithTTL(ctx context.Context, key string, ttlSeconds int) (int, error) {
	n, err := s.rds.IncrCtx(ctx, key)
	if err != nil {
		return 0, err
	}
	if n == 1 {
		if err := s.rds.ExpireCtx(ctx, key, ttlSeconds); err != nil {
			return 0, err
		}
	}
	return int(n), nil
}
