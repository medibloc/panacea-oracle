package server

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/medibloc/panacea-oracle/server/middleware"
	"github.com/medibloc/panacea-oracle/service"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	*http.Server
}

func New(svc *service.Service) *Server {
	router := mux.NewRouter()
	router.HandleFunc("/v0/data-deal/deals/{dealId}/data", svc.ValidateData).Methods("POST")
	router.HandleFunc("/v0/data-deal/secret-key", svc.GetSecretKey).Methods("GET")

	mw := middleware.NewJWTAuthMiddleware(svc.QueryClient())
	router.Use(mw.Middleware)

	return &Server{
		&http.Server{
			Handler:      router,
			Addr:         svc.Config().ListenAddr,
			WriteTimeout: time.Duration(svc.Config().API.WriteTimeout) * time.Second,
			ReadTimeout:  time.Duration(svc.Config().API.ReadTimeout) * time.Second,
		},
	}
}

func (srv *Server) Run() error {
	log.Infof("HTTP server is started: %s", srv.Addr)
	return srv.ListenAndServe()
}
