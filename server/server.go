package server

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"github.com/medibloc/panacea-oracle/server/middleware"
	"github.com/medibloc/panacea-oracle/server/service/datadeal"
	"github.com/medibloc/panacea-oracle/server/service/key"
	"github.com/medibloc/panacea-oracle/server/service/status"
	"github.com/medibloc/panacea-oracle/service"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/netutil"
)

type Server struct {
	*http.Server
	maxConnections int
}

func New(svc service.Service) *Server {
	router := mux.NewRouter()
	conf := svc.Config()

	limitMiddleware := middleware.NewLimitMiddleware(conf.API.MaxRequestBodySize)
	jwtAuthMiddleware := middleware.NewJWTAuthMiddleware(svc.QueryClient())

	dealRouter := router.PathPrefix("/v0/data-deal").Subrouter()
	dealRouter.Use(
		limitMiddleware.Middleware,
		jwtAuthMiddleware.Middleware,
	)

	datadeal.RegisterHandlers(svc, dealRouter)
	key.RegisterHandlers(svc, dealRouter)

	status.RegisterHandlers(svc, router)

	return &Server{
		&http.Server{
			Handler:      router,
			Addr:         conf.API.ListenAddr,
			WriteTimeout: time.Duration(conf.API.WriteTimeout) * time.Second,
			ReadTimeout:  time.Duration(conf.API.ReadTimeout) * time.Second,
		},
		conf.API.MaxConnections,
	}
}

func (srv *Server) Run() error {
	listenURL, err := url.Parse(srv.Addr)
	if err != nil {
		return fmt.Errorf("failed to parse api address. address: %s, %w", srv.Addr, err)
	}
	lis, err := net.Listen(listenURL.Scheme, listenURL.Host)
	if err != nil {
		return err
	}

	log.Infof("HTTP server is started: %s", srv.Addr)
	return srv.Serve(netutil.LimitListener(lis, srv.maxConnections))
}
