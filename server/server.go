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

func Serve(svc service.Service) ([]Server, chan error) {
	cfg := svc.Config()

	var servers []Server

	errCh := make(chan error)
	svr := rpc.NewGrpcServer(svc)
	servers = append(servers, svr)
	go runServer(svr, errCh)

	log.Infof("API enabled: %v", cfg.API.Enabled)
	if cfg.API.Enabled {
		svr, err := rpc.NewGatewayServer(cfg)
		if err != nil {
			errCh <- err
		} else {
			servers = append(servers, svr)
			go runServer(svr, errCh)
		}
	}

	return servers, errCh
}

func runServer(svr Server, errCh chan error) {
	if err := svr.Run(); err != nil {
		errCh <- err
	}
	defer svr.Close()
}
