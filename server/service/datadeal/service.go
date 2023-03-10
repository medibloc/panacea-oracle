package datadeal

import (
	"github.com/medibloc/panacea-oracle/validation"
	"net/http"

	"github.com/gorilla/mux"
	serverservice "github.com/medibloc/panacea-oracle/server/service"
)

type dataDealService struct {
	serverservice.Service
	schema *validation.JSONSchema
}

func RegisterHandlers(svc serverservice.Service, router *mux.Router) {
	s := &dataDealService{
		Service: svc,
		schema:  validation.NewJSONSchema(),
	}

	router.HandleFunc("/deals/{dealId}/data", s.ValidateData).Methods(http.MethodPost)
}
