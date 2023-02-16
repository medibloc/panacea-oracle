package status

import (
	"context"

	status "github.com/medibloc/panacea-oracle/pb/status/v0"
)

func (s *statusService) GetStatus(ctx context.Context, req *status.GetStatusRequest) (*status.GetStatusResponse, error) {
	return &status.GetStatusResponse{
		OracleAccountAddress: s.GetOracleAcc().GetAddress(),
		Api: &status.StatusAPI{
			Enabled:    s.GetConfig().API.Enabled,
			ListenAddr: s.GetConfig().API.ListenAddr,
		},
		Grpc: &status.StatusGRPC{
			ListenAddr: s.GetConfig().GRPC.ListenAddr,
		},
		EnclaveInfo: &status.StatusEnclaveInfo{
			ProductId: s.GetEnclaveInfo().ProductID,
			UniqueId:  s.GetEnclaveInfo().UniqueIDHex(),
		},
	}, nil
}
