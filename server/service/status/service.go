package status

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	status "github.com/medibloc/panacea-oracle/pb/status/v0"
	serverservice "github.com/medibloc/panacea-oracle/server/service"
	"google.golang.org/grpc"
)

type statusService struct {
	status.UnimplementedStatusServiceServer

	serverservice.Service
}

func RegisterService(svc serverservice.Service, svr *grpc.Server) {
	status.RegisterStatusServiceServer(svr, &statusService{Service: svc})
}

func RegisterServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return status.RegisterStatusServiceHandler(ctx, mux, conn)
}
