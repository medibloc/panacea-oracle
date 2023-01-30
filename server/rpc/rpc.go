package rpc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/medibloc/panacea-oracle/server/service/status"
	"github.com/medibloc/panacea-oracle/service"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func Serve(svc service.Service, errCh chan error) error {
	if err := serverGRPC(svc, errCh); err != nil {
		return fmt.Errorf("failed to serve gRPC server: %w", err)
	}

	if err := serverREST(svc, errCh); err != nil {
		return fmt.Errorf("failed to serve REST server: %w", err)
	}

	return nil
}

func serverGRPC(svc service.Service, errCh chan error) error {
	grpcListenURL, err := url.Parse(svc.Config().GRPC.ListenAddr)
	if err != nil {
		return err
	}

	svr := grpc.NewServer()
	status.RegisterService(svr)

	lis, err := net.Listen(grpcListenURL.Scheme, grpcListenURL.Host)
	if err != nil {
		return fmt.Errorf("failed to listen port for RPC: %w", err)
	}

	go func() {
		log.Infof("gRPC server listening at %s...", grpcListenURL.Host)
		if err := svr.Serve(lis); err != nil {
			errCh <- err
			return
		}
	}()

	return nil
}

func serverREST(svc service.Service, errCh chan error) error {
	cfg := svc.Config()
	grpcListenURL, err := url.Parse(cfg.GRPC.ListenAddr)
	if err != nil {
		return err
	}

	restListenURL, err := url.Parse(cfg.API.ListenAddr)
	if err != nil {
		return err
	}
	restListenAddr := restListenURL.Host

	ctx := context.Background()

	grpcEndpoint := fmt.Sprintf("127.0.0.1:%s", grpcListenURL.Port())

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()

	conn, err := grpc.DialContext(
		ctx,
		grpcEndpoint,
		grpc.WithBlock(),
		grpc.WithInsecure(),
	)

	if err != nil {
		return err
	}

	if err := status.RegisterHandler(ctx, mux, conn); err != nil {
		return err
	}

	go func() {
		log.Infof("REST  server listening at %s...", restListenAddr)
		if err := http.ListenAndServe(restListenAddr, mux); err != nil {
			errCh <- err
			return
		}
	}()
	return nil
}
