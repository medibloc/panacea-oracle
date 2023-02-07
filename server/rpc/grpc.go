package rpc

import (
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/medibloc/panacea-oracle/server/rpc/interceptor/auth"
	"github.com/medibloc/panacea-oracle/server/rpc/interceptor/query"
	serverservice "github.com/medibloc/panacea-oracle/server/service"
	"github.com/medibloc/panacea-oracle/server/service/datadeal"
	"github.com/medibloc/panacea-oracle/server/service/key"
	"github.com/medibloc/panacea-oracle/server/service/status"
	"github.com/medibloc/panacea-oracle/service"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/netutil"
	"google.golang.org/grpc"
)

type GrpcServer struct {
	grpcServer *grpc.Server
	svc        service.Service
}

func NewGrpcServer(svc service.Service) *GrpcServer {
	unaryInterceptor, streamInterceptor := createInterceptors(svc)

	grpcSvr := grpc.NewServer(
		unaryInterceptor,
		streamInterceptor,
		grpc.ConnectionTimeout(time.Duration(svc.Config().GRPC.ConnectionTimeout)*time.Second),
	)

	return &GrpcServer{
		grpcSvr,
		svc,
	}
}

func createInterceptors(svc service.Service) (grpc.ServerOption, grpc.ServerOption) {
	jwtAuthInterceptor := auth.NewJWTAuthInterceptor(svc.QueryClient())
	queryInterceptor := query.NewQueryInterceptor(svc.QueryClient())

	return grpc.ChainUnaryInterceptor(
			jwtAuthInterceptor.UnaryServerInterceptor(),
			queryInterceptor.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			jwtAuthInterceptor.StreamServerInterceptor(),
			queryInterceptor.StreamServerInterceptor(),
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

func (s *GrpcServer) registerServices(registerServices ...func(serverservice.Service, *grpc.Server)) {
	log.Info("Register grpc services")
	for _, registerService := range registerServices {
		registerService(s.svc, s.grpcServer)
	}
}

func (s *GrpcServer) listenAndServe() error {
	grpcListenURL, err := url.Parse(s.svc.Config().GRPC.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to parsing rest URL: %w", err)
	}

	log.Infof("gRPC server is started: %s", grpcListenURL.Host)
	lis, err := net.Listen(grpcListenURL.Scheme, grpcListenURL.Host)
	if err != nil {
		return fmt.Errorf("failed to listen port for RPC: %w", err)
	}

	lis = netutil.LimitListener(lis, 100)

	return s.grpcServer.Serve(lis)
}
