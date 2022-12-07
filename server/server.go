package server

import (
	"github.com/gorilla/mux"
	"github.com/medibloc/panacea-oracle/service"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type Server struct {
	*http.Server
}

func New(svc *service.Service) *Server {
	router := mux.NewRouter()
	router.HandleFunc("/v0/data-deal/deals/{dealId}/data", svc.ValidateData).Methods("POST")

	return &Server{
		&http.Server{
			Handler:      router,
			Addr:         svc.Config().ListenAddr,
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		},
	}
}

func (srv *Server) Run() error {
	log.Infof("HTTP server is started: %s", srv.Addr)
	return srv.ListenAndServe()
}
