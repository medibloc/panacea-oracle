package rpc

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/medibloc/panacea-oracle/server/service/datadeal"
	"github.com/medibloc/panacea-oracle/server/service/key"
	"github.com/medibloc/panacea-oracle/server/service/status"
	"github.com/medibloc/panacea-oracle/service"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
)

type gatewayServer struct {
	grpcConn *grpc.ClientConn
	svr      *http.Server
	service  service.Service
}

func NewGatewayServer(svc service.Service) *gatewayServer {
	return &gatewayServer{
		service: svc,
	}
}

func (s *gatewayServer) Run() error {
	log.Info("Running the API server")

	if err := s.createAndSetGrpcConnection(); err != nil {
		return err
	}

	mux := runtime.NewServeMux()
	if err := s.registerServiceHandler(mux); err != nil {
		return err
	}

	return s.listenAndServe(mux)
}

func (s *gatewayServer) createAndSetGrpcConnection() error {
	grpcListenURL, err := url.Parse(s.service.Config().GRPC.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to parsing rest URL: %w", err)
	}
	port := grpcListenURL.Port()

	grpcEndpoint := fmt.Sprintf("127.0.0.1:%s", port)

	log.Infof("Dial gateway to gRPC server > %s", grpcEndpoint)

	conn, err := grpc.DialContext(
		context.Background(),
		grpcEndpoint,
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff:           backoff.DefaultConfig,
			MinConnectTimeout: time.Duration(s.service.Config().API.GrpcConnectionTimeout) * time.Second,
		}),
		grpc.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("failed to generate grpc connection. %w", err)
	}
	s.grpcConn = conn

	return nil
}

func (s *gatewayServer) registerServiceHandler(mux *runtime.ServeMux) error {
	return s.registerServiceHandlers(
		mux,
		datadeal.RegisterServiceHandler,
		key.RegisterServiceHandler,
		status.RegisterServiceHandler,
	)
}

func (s *gatewayServer) registerServiceHandlers(mux *runtime.ServeMux, handlers ...func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error) error {
	log.Info("Register service handlers")
	ctx := context.Background()

	for _, handler := range handlers {
		if err := handler(ctx, mux, s.grpcConn); err != nil {
			return err
		}
	}
	return nil
}

func (s *gatewayServer) listenAndServe(mux *runtime.ServeMux) error {
	cfg := s.service.Config()

	restListenURL, err := url.Parse(cfg.API.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to parsing rest URL: %w", err)
	}
	log.Infof("API server is started: %s", restListenURL.Host)

	svr := &http.Server{
		Handler:      mux,
		Addr:         restListenURL.Host,
		WriteTimeout: time.Duration(cfg.API.WriteTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.API.ReadTimeout) * time.Second,
	}
	s.svr = svr

	return svr.ListenAndServe()
}

func (s *gatewayServer) Close() error {
	log.Info("Close API server")
	if err := s.svr.Close(); err != nil {
		log.Warnf("faled to close gateway http server. %v", err)
	}

	if s.grpcConn != nil {
		if err := s.grpcConn.Close(); err != nil {
			log.Warnf("faled to close gateway grpc connection. %v", err)
		}
	}

	return nil
}
