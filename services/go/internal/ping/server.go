package ping

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	pingv1 "github.com/snowaa-desigram/backend/services/go/gen/desigram/ping/v1"
)

// Server — заглушка PingService.
type Server struct {
	pingv1.UnimplementedPingServiceServer
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Ping(ctx context.Context, req *pingv1.PingRequest) (*pingv1.PingResponse, error) {
	logx.WithContext(ctx).Infof("ping: %s", req.GetMessage())

	return &pingv1.PingResponse{
		Message: "pong: " + req.GetMessage(),
		Service: "ping",
	}, nil
}
