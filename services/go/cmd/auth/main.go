package main

import (
	"context"
	"errors"
	"flag"
	"time"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"

	"github.com/snowaa-desigram/backend/services/go/internal/auth"
)

var (
	configFile = flag.String("f", "etc/auth.yaml", "config file")
	migrate    = flag.Bool("migrate", false, "apply DB migrations and exit (compose: auth-migrate)")
)

func main() {
	flag.Parse()

	var c auth.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	// php-jwt в core требует ключ HS256 не короче 32 байт — проверяем на старте, а не в проде на первом запросе
	if len(c.Auth.AccessSecret) < 32 {
		logx.Must(errors.New("JWT_SECRET must be at least 32 bytes"))
	}

	db, err := auth.OpenDB(c.DB.DataSource)
	logx.Must(err)

	if *migrate {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		logx.Must(auth.Migrate(ctx, db))
		logx.Info("migrations applied")
		return
	}

	rds := redis.MustNewRedis(c.Redis)
	mailer, err := auth.NewSMTPMailer(c.Smtp)
	logx.Must(err)

	svc := auth.NewService(
		auth.NewGormUserStore(db, rds),
		auth.NewGormRefreshTokenStore(db),
		auth.NewRedisCodeStore(rds),
		mailer,
		auth.NewTokenIssuer(c.Auth.AccessSecret, time.Duration(c.Auth.AccessExpire)*time.Second, time.Now),
		time.Now,
		c.Options(),
	)

	server := rest.MustNewServer(c.RestConf, rest.WithUnauthorizedCallback(auth.UnauthorizedCallback))
	defer server.Stop()

	auth.RegisterRoutes(server, auth.NewHandler(svc), c.Auth.AccessSecret)
	server.Start()
}
