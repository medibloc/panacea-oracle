package status

import (
	"context"

	v0 "github.com/medibloc/panacea-oracle/pb/status/v0"
	"google.golang.org/grpc"
)

type statusServiceServer struct {
	v0.UnimplementedStatusServiceServer
}

func (s *statusServiceServer) GetStatus(ctx context.Context, req *v0.GetStatusRequest) (*v0.GetStatusResponse, error) {
	return &v0.GetStatusResponse{
		OracleAccountAddress: "no",
		Api: &v0.StatusAPI{
			ListenAddr: "",
		},
		EnclaveInfo: &v0.StatusEnclaveInfo{
			ProductId: []byte("asdfasdfsadf"),
			UniqueId:  "45678",
		},
	}, nil
}

func RegisterService(svr *grpc.Server) {
	v0.RegisterStatusServiceServer(svr, &statusServiceServer{})
}
