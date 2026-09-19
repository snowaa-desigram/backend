package transport

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	pingv1 "github.com/snowaa-desigram/backend/services/go/gen/desigram/ping/v1"
	"github.com/snowaa-desigram/backend/services/go/internal/ping/service"
)

// Server — gRPC-вход PingService: pb → Service → pb.
type Server struct {
	pingv1.UnimplementedPingServiceServer
	svc *service.Service
}

func NewServer(svc *service.Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) Ping(ctx context.Context, req *pingv1.PingRequest) (*pingv1.PingResponse, error) {
	logx.WithContext(ctx).Infof("ping: %s", req.GetMessage())

	return &pingv1.PingResponse{
		Message: s.svc.Ping(req.GetMessage()),
		Service: "ping",
	}, nil
}
