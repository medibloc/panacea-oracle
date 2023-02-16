package datadeal

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	datadeal "github.com/medibloc/panacea-oracle/pb/datadeal/v0"
	serverservice "github.com/medibloc/panacea-oracle/server/service"
	"google.golang.org/grpc"
)

type dataDealServiceServer struct {
	datadeal.UnimplementedDataDealServiceServer

	serverservice.Service
}

func RegisterService(svc serverservice.Service, svr *grpc.Server) {
	datadeal.RegisterDataDealServiceServer(svr, &dataDealServiceServer{
		Service: svc,
	})
}

func RegisterServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return datadeal.RegisterDataDealServiceHandler(ctx, mux, conn)
}
