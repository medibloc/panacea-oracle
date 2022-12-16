package key

import (
	"net/http"

	"github.com/gorilla/mux"
	serverservice "github.com/medibloc/panacea-oracle/server/service"
)

type combinedKeyService struct {
	serverservice.Service
}

func RegisterHandlers(svc serverservice.Service, router *mux.Router) {
	s := &combinedKeyService{svc}

	router.HandleFunc("/secret-key", s.GetSecretKey).Methods(http.MethodGet)
}
