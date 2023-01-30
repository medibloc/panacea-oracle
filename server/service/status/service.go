package status

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	status "github.com/medibloc/panacea-oracle/pb/status/v0"
	serverservice "github.com/medibloc/panacea-oracle/server/service"
	"google.golang.org/grpc"
)

type statusService struct {
	serverservice.Service
}

func RegisterHandlers(svc serverservice.Service, router *mux.Router) {
	s := &statusService{svc}

	router.HandleFunc("/v0/status", s.GetStatus).Methods(http.MethodGet)
}

type statusServiceServer struct {
	status.UnimplementedStatusServiceServer
}

func RegisterService(svr *grpc.Server) {
	status.RegisterStatusServiceServer(svr, &statusServiceServer{})
}

func RegisterHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return status.RegisterStatusServiceHandler(ctx, mux, conn)
}
