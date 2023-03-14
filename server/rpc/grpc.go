package rpc

import (
	"fmt"
	"net"

	"github.com/medibloc/panacea-oracle/server/rpc/interceptor/auth"
	"github.com/medibloc/panacea-oracle/server/rpc/interceptor/limit"
	"github.com/medibloc/panacea-oracle/server/service/datadeal"
	"github.com/medibloc/panacea-oracle/server/service/key"
	"github.com/medibloc/panacea-oracle/server/service/status"
	"github.com/medibloc/panacea-oracle/service"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/netutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type GrpcServer struct {
	grpcServer *grpc.Server
	svc        service.Service
}

func NewGrpcServer(svc service.Service) *GrpcServer {
	cfg := svc.Config().GRPC

	unaryInterceptor, streamInterceptor := createInterceptors(svc)

	grpcSvr := grpc.NewServer(
		unaryInterceptor,
		streamInterceptor,
		grpc.ConnectionTimeout(cfg.ConnectionTimeout),
		grpc.MaxConcurrentStreams(uint32(cfg.MaxConcurrentStreams)),
		grpc.MaxRecvMsgSize(cfg.MaxRecvMsgSize),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     cfg.KeepaliveMaxConnectionIdle,
			MaxConnectionAge:      cfg.KeepaliveMaxConnectionAge,
			MaxConnectionAgeGrace: cfg.KeepaliveMaxConnectionAgeGrace,
			Time:                  cfg.KeepaliveTime,
			Timeout:               cfg.KeepaliveTimeout,
		}),
	)

	return &GrpcServer{
		grpcSvr,
		svc,
	}
}

func createInterceptors(svc service.Service) (grpc.ServerOption, grpc.ServerOption) {
	jwtAuthInterceptor := auth.NewJWTAuthInterceptor(svc.QueryClient())
	rateLimitInterceptor := limit.NewRateLimitInterceptor(svc.Config().GRPC)

	return grpc.ChainUnaryInterceptor(
			rateLimitInterceptor.UnaryServerInterceptor(),
			jwtAuthInterceptor.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			rateLimitInterceptor.StreamServerInterceptor(),
			jwtAuthInterceptor.StreamServerInterceptor(),
		)

}

func (s *GrpcServer) Run() error {
	log.Info("Running the gRPC server")

	s.registerServices(
		datadeal.RegisterService,
		key.RegisterService,
		status.RegisterService,
	)

	return s.listenAndServe()
}

func (s *GrpcServer) Close() error {
	log.Info("Close gRPC server")
	s.grpcServer.GracefulStop()

	return nil
}

func (s *GrpcServer) registerServices(registerServices ...func(service.Service, *grpc.Server)) {
	log.Info("Register grpc services")
	for _, registerService := range registerServices {
		registerService(s.svc, s.grpcServer)
	}
}

func (s *GrpcServer) listenAndServe() error {
	cfg := s.svc.Config().GRPC

	log.Infof("gRPC server is started: %s", cfg.ListenAddr)

	lis, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen port for RPC: %w", err)
	}

	return s.grpcServer.Serve(netutil.LimitListener(lis, cfg.MaxConnections))
}
