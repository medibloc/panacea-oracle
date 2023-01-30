package rpc

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	v0 "github.com/medibloc/panacea-oracle/pb/status/v0"
	"github.com/medibloc/panacea-oracle/server/service/status"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func Serve(rpcPort, restPort int, errCh chan error) error {
	err := serverGRPC(rpcPort, errCh)
	if err != nil {
		return fmt.Errorf("failed to serve gRPC server: %w", err)
	}

	if err := serverREST(restPort, rpcPort, errCh); err != nil {
		return fmt.Errorf("failed to serve REST server: %w", err)
	}

	return err
}

func serverGRPC(port int, errCh chan error) error {
	svr := grpc.NewServer()
	status.RegisterService(svr)

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen port for RPC: %w", err)
	}

	go func() {
		log.Infof("gRPC server listening at %d...", port)
		if err := svr.Serve(lis); err != nil {
			errCh <- err
			return
		}
	}()

	return nil
}

func serverREST(port, rpcPort int, errCh chan error) error {
	ctx := context.Background()

	rpcEndpoint := fmt.Sprintf("127.0.0.1:%d", rpcPort)

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()

	conn, err := grpc.DialContext(
		ctx,
		rpcEndpoint,
		grpc.WithBlock(),
		grpc.WithInsecure(),
	)

	if err != nil {
		return err
	}

	if err := v0.RegisterStatusServiceHandler(ctx, mux, conn); err != nil {
		return err
	}

	go func() {
		log.Infof("REST  server listening at %d...", port)
		if err := http.ListenAndServe("0.0.0.0:8080", mux); err != nil {
			errCh <- err
			return
		}
	}()
	return nil
}
