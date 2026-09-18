package main

import (
	"flag"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pingv1 "github.com/snowaa-desigram/backend/services/go/gen/desigram/ping/v1"
	"github.com/snowaa-desigram/backend/services/go/internal/ping"
)

var configFile = flag.String("f", "etc/ping.yaml", "config file")

func main() {
	flag.Parse()

	var c ping.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pingv1.RegisterPingServiceServer(grpcServer, ping.NewServer())

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	s.Start()
}
