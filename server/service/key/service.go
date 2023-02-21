package key

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	key "github.com/medibloc/panacea-oracle/pb/key/v0"
	"github.com/medibloc/panacea-oracle/service"
	"google.golang.org/grpc"
)

var _ service.Service = &secretKeyService{}

type secretKeyService struct {
	key.UnimplementedKeyServiceServer

	service.Service
}

func RegisterService(svc service.Service, svr *grpc.Server) {
	key.RegisterKeyServiceServer(svr, &secretKeyService{
		Service: svc,
	})
}

func RegisterServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return key.RegisterKeyServiceHandler(ctx, mux, conn)
}
