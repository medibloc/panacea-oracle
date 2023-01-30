package status

import (
	"context"

	v0 "github.com/medibloc/panacea-oracle/pb/status/v0"
)

func (s *statusServiceServer) GetStatus(ctx context.Context, req *v0.GetStatusRequest) (*v0.GetStatusResponse, error) {
	return &v0.GetStatusResponse{
		OracleAccountAddress: "no",
		Api:                  &v0.StatusAPI{
			//ListenAddr: s.Config().API.ListenAddr,
		},
		EnclaveInfo: &v0.StatusEnclaveInfo{
			/*ProductId: s.EnclaveInfo().ProductID,
			UniqueId:  s.EnclaveInfo().UniqueIDHex(),*/
		},
	}, nil
}
