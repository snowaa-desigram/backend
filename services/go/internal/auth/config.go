package auth

import (
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

// Config сервиса: rest.RestConf даёт Host/Port, Mode, Log, Prometheus, Telemetry, Timeout и т.д.
type Config struct {
	rest.RestConf

	// Имя продукта для писем (APP_NAME из group_vars app_name).
	AppName string `json:",default=Gram Designer"`

	// JWT: секрет общий с core (JWT_SECRET), AccessExpire — секунды.
	Auth struct {
		AccessSecret string
		AccessExpire int64 `json:",default=900"`
	}
	// Срок жизни refresh-токена, секунды.
	RefreshExpire int64 `json:",default=2592000"`

	DB struct {
		DataSource string
	}
	// Один Redis: кеш для sqlc (пользователи) и коды/счётчики попыток.
	Redis redis.RedisConf
	Smtp  SMTPConfig

	Code struct {
		TTL         time.Duration `json:",default=10m"`
		Cooldown    time.Duration `json:",default=60s"`
		MaxAttempts int           `json:",default=5"`
	}
	Login struct {
		MaxFailures int           `json:",default=10"`
		Window      time.Duration `json:",default=15m"`
	}
}

// Options — параметры логики, вынесены из Config, чтобы Service не зависел от go-zero conf.
func (c Config) Options() Options {
	return Options{
		AppName:          c.AppName,
		AccessTTL:        time.Duration(c.Auth.AccessExpire) * time.Second,
		RefreshTTL:       time.Duration(c.RefreshExpire) * time.Second,
		CodeTTL:          c.Code.TTL,
		CodeCooldown:     c.Code.Cooldown,
		CodeMaxAttempts:  c.Code.MaxAttempts,
		LoginMaxFailures: c.Login.MaxFailures,
		LoginWindow:      c.Login.Window,
	}
}
