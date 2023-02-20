package status

import (
	"context"

	status "github.com/medibloc/panacea-oracle/pb/status/v0"
)

func (s *statusService) GetStatus(ctx context.Context, req *status.GetStatusRequest) (*status.GetStatusResponse, error) {
	return &status.GetStatusResponse{
		OracleAccountAddress: s.OracleAcc().GetAddress(),
		Api: &status.StatusAPI{
			Enabled:    s.Config().API.Enabled,
			ListenAddr: s.Config().API.ListenAddr,
		},
		Grpc: &status.StatusGRPC{
			ListenAddr: s.Config().GRPC.ListenAddr,
		},
		EnclaveInfo: &status.StatusEnclaveInfo{
			ProductId: s.EnclaveInfo().ProductID,
			UniqueId:  s.EnclaveInfo().UniqueIDHex(),
		},
	}, nil
}
