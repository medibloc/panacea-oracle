package key

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	key "github.com/medibloc/panacea-oracle/pb/key/v0"
	serverservice "github.com/medibloc/panacea-oracle/server/service"
	"google.golang.org/grpc"
)

var _ serverservice.Service = &combinedKeyService{}

type combinedKeyService struct {
	key.UnimplementedKeyServiceServer
	serverservice.Service
}

func RegisterService(svc serverservice.Service, svr *grpc.Server) {
	key.RegisterKeyServiceServer(svr, &combinedKeyService{
		Service: svc,
	})
}

func RegisterServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return key.RegisterKeyServiceHandler(ctx, mux, conn)
}
