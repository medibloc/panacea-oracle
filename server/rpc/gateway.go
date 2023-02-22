package rpc

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/medibloc/panacea-oracle/config"
	"github.com/medibloc/panacea-oracle/server/service/datadeal"
	"github.com/medibloc/panacea-oracle/server/service/key"
	"github.com/medibloc/panacea-oracle/server/service/status"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
)

type GatewayServer struct {
	*http.Server
	grpcConn *grpc.ClientConn
}

func NewGatewayServer(cfg *config.Config) (*GatewayServer, error) {
	mux := runtime.NewServeMux()

	conn, err := createGrpcConnection(cfg)
	if err != nil {
		return nil, err
	}

	if err := registerServiceHandlers(
		mux,
		conn,
		datadeal.RegisterServiceHandler,
		key.RegisterServiceHandler,
		status.RegisterServiceHandler,
	); err != nil {
		return nil, fmt.Errorf("failed to register service handlers: %w", err)
	}

	restListenURL, err := url.Parse(cfg.API.ListenAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to parsing rest URL: %w", err)
	}

	return &GatewayServer{
		Server: &http.Server{
			Handler:      mux,
			Addr:         restListenURL.Host,
			WriteTimeout: cfg.API.WriteTimeout,
			ReadTimeout:  cfg.API.ReadTimeout,
		},
		grpcConn: conn,
	}, nil
}

func (s *GatewayServer) Run() error {
	log.Infof("API server is started: %s", s.Addr)

	return s.ListenAndServe()
}

func createGrpcConnection(cfg *config.Config) (*grpc.ClientConn, error) {
	grpcListenURL, err := url.Parse(cfg.GRPC.ListenAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to parsing rest URL: %w", err)
	}

	log.Infof("Dial gateway to gRPC server > %s", grpcListenURL.Host)

	return grpc.DialContext(
		context.Background(),
		grpcListenURL.Host,
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff:           backoff.DefaultConfig,
			MinConnectTimeout: cfg.API.GrpcConnectTimeout,
		}),
		grpc.WithInsecure(),
	)
}

func registerServiceHandlers(mux *runtime.ServeMux, conn *grpc.ClientConn, handlers ...func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error) error {
	log.Info("Register service handlers")
	ctx := context.Background()

	for _, handler := range handlers {
		if err := handler(ctx, mux, conn); err != nil {
			return err
		}
	}
	return nil
}

func (s *GatewayServer) Close() error {
	log.Info("Close API server")

	ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := s.Server.Shutdown(ctxTimeout); err != nil {
		log.Warnf("failed to close gateway http server. %v", err)
	}

	if s.grpcConn != nil {
		if err := s.grpcConn.Close(); err != nil {
			log.Warnf("failed to close gateway grpc connection. %v", err)
		}
	}

	return nil
}
