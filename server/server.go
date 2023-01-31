package server

import (
	"github.com/medibloc/panacea-oracle/server/rpc"
	"github.com/medibloc/panacea-oracle/service"
	log "github.com/sirupsen/logrus"
)

type Server interface {
	Run() error

	Close() error
}

func Serve(svc service.Service) chan error {
	cfg := svc.Config()

	errCh := make(chan error)
	log.Infof("gRPC enabled: %v", cfg.GRPC.Enabled)
	if cfg.GRPC.Enabled {
		runServer(rpc.NewGrpcServer(svc), errCh)
	}

	log.Infof("API enabled: %v", cfg.API.Enabled)
	if cfg.API.Enabled {
		if !cfg.GRPC.Enabled {
			log.Warnf("gRPC server is not running. The API server needs to run a gRPC server.")
		} else {
			runServer(rpc.NewGatewayServer(svc), errCh)
		}
	}

	return errCh
}

func runServer(svr Server, errCh chan error) {
	go func() {
		if err := svr.Run(); err != nil {
			errCh <- err
		}
		defer svr.Close()
	}()
}
