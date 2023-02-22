package status

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	status "github.com/medibloc/panacea-oracle/pb/status/v0"
	"github.com/medibloc/panacea-oracle/service"
	"google.golang.org/grpc"
)

type statusService struct {
	status.UnimplementedStatusServiceServer

	service.Service
}

func RegisterService(svc service.Service, svr *grpc.Server) {
	status.RegisterStatusServiceServer(svr, &statusService{Service: svc})
}

func RegisterServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return status.RegisterStatusServiceHandler(ctx, mux, conn)
}
