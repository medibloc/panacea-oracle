package rpc

import (
	"context"
	"fmt"
	"golang.org/x/net/netutil"
	"net"
	"net/http"
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
	grpcConn       *grpc.ClientConn
	maxConnections int
}

func NewGatewayServer(conf *config.Config) (*GatewayServer, error) {
	mux := runtime.NewServeMux()

	conn, err := createGrpcConnection(conf)
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

	return &GatewayServer{
		Server: &http.Server{
			Handler:      appendPreHandlers(mux, conf),
			Addr:         conf.API.ListenAddr,
			WriteTimeout: conf.API.WriteTimeout,
			ReadTimeout:  conf.API.ReadTimeout,
		},
		grpcConn:       conn,
		maxConnections: conf.API.MaxConnections,
	}, nil
}

func createGrpcConnection(conf *config.Config) (*grpc.ClientConn, error) {
	log.Infof("Dial gateway to gRPC server > %s", conf.GRPC.ListenAddr)

	return grpc.DialContext(
		context.Background(),
		conf.GRPC.ListenAddr,
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff:           backoff.DefaultConfig,
			MinConnectTimeout: conf.API.GrpcConnectTimeout,
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

// appendPreHandlers implements handlers that should be processed before every request
func appendPreHandlers(handler http.Handler, conf *config.Config) http.Handler {
	return appendLimitRequestBodySizeHandler(handler, conf.API.MaxRequestBodySize)
}

// appendLimitRequestBodySizeHandler limits the request body size.
// This is done by first constraining to the ContentLength of the request header,
// and then reading the actual Body to constraint it.
func appendLimitRequestBodySizeHandler(handler http.Handler, maxRequestBodySize int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > maxRequestBodySize {
			http.Error(w, "request body too large", http.StatusBadRequest)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
		defer r.Body.Close()

		handler.ServeHTTP(w, r)
	})
}

func (s *GatewayServer) Run() error {
	lis, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	log.Infof("API server is started: %s", s.Addr)
	return s.Serve(netutil.LimitListener(lis, s.maxConnections))
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
