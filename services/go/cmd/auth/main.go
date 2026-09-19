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
	"github.com/snowaa-desigram/backend/services/go/internal/auth/adapter"
	"github.com/snowaa-desigram/backend/services/go/internal/auth/service"
	"github.com/snowaa-desigram/backend/services/go/internal/auth/store"
	"github.com/snowaa-desigram/backend/services/go/internal/auth/transport"
)

var (
	configFile = flag.String("f", "etc/auth.yaml", "config file")
	migrate    = flag.Bool("migrate", false, "apply DB migrations and exit (compose: auth-migrate)")
)

func main() {
	flag.Parse()

	var c auth.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	if len(c.Auth.AccessSecret) < 32 {
		logx.Must(errors.New("JWT_SECRET must be at least 32 bytes"))
	}

	db, err := store.OpenDB(c.DB.DataSource)
	logx.Must(err)

	if *migrate {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		logx.Must(store.Migrate(ctx, db))
		logx.Info("migrations applied")
		return
	}

	rds := redis.MustNewRedis(c.Redis)
	mailer, err := adapter.NewSMTPMailer(c.Smtp)
	logx.Must(err)

	svc := service.NewService(
		store.NewGormUserStore(db, rds),
		store.NewGormRefreshTokenStore(db),
		store.NewRedisCodeStore(rds),
		mailer,
		service.NewTokenIssuer(c.Auth.AccessSecret, time.Duration(c.Auth.AccessExpire)*time.Second, time.Now),
		time.Now,
		c.Options(),
	)

	server := rest.MustNewServer(c.RestConf, rest.WithUnauthorizedCallback(transport.UnauthorizedCallback))
	defer server.Stop()

	transport.RegisterRoutes(server, transport.NewHandler(svc), c.Auth.AccessSecret)
	server.Start()
}
