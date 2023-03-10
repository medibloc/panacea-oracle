package server

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/medibloc/panacea-oracle/server/middleware"
	"github.com/medibloc/panacea-oracle/server/service/datadeal"
	"github.com/medibloc/panacea-oracle/server/service/key"
	"github.com/medibloc/panacea-oracle/server/service/status"
	"github.com/medibloc/panacea-oracle/service"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	*http.Server
}

func New(svc service.Service) *Server {
	router := mux.NewRouter()

	jwtAuthMiddleware := middleware.NewJWTAuthMiddleware(svc.QueryClient())

	dealRouter := router.PathPrefix("/v0/data-deal").Subrouter()
	dealRouter.Use(
		jwtAuthMiddleware.Middleware,
	)

	datadeal.RegisterHandlers(svc, dealRouter)
	key.RegisterHandlers(svc, dealRouter)

	status.RegisterHandlers(svc, router)

	return &Server{
		&http.Server{
			Handler:      router,
			Addr:         svc.Config().API.ListenAddr,
			WriteTimeout: time.Duration(svc.Config().API.WriteTimeout) * time.Second,
			ReadTimeout:  time.Duration(svc.Config().API.ReadTimeout) * time.Second,
		},
	}
}

func (srv *Server) Run() error {
	log.Infof("HTTP server is started: %s", srv.Addr)
	return srv.ListenAndServe()
}
