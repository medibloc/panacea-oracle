package datadeal

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	datadeal "github.com/medibloc/panacea-oracle/pb/datadeal/v0"
	"github.com/medibloc/panacea-oracle/service"
	"github.com/medibloc/panacea-oracle/validation"
	"google.golang.org/grpc"
)

type dataDealServiceServer struct {
	datadeal.UnimplementedDataDealServiceServer

	service.Service
	schema *validation.JSONSchema
}

func RegisterService(svc service.Service, svr *grpc.Server) {
	datadeal.RegisterDataDealServiceServer(svr, &dataDealServiceServer{
		Service: svc,
		schema:  validation.NewJSONSchema(),
	})
}

func RegisterServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return datadeal.RegisterDataDealServiceHandler(ctx, mux, conn)
}
