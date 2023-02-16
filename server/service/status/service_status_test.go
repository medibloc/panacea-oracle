package status

import (
	"context"
	"testing"

	"github.com/medibloc/panacea-oracle/config"
	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/sgx"
	"github.com/medibloc/panacea-oracle/types/test_utils/mocks"
	"github.com/stretchr/testify/require"
)

func TestGetStatus(t *testing.T) {
	mnemonic, err := crypto.NewMnemonic()
	require.NoError(t, err)
	oracleAcc, err := panacea.NewOracleAccount(mnemonic, 0, 0)
	require.NoError(t, err)

	conf := config.DefaultConfig()
	svc := &mocks.MockService{
		OracleAccount: oracleAcc,
		Config:        conf,
		EnclaveInfo: sgx.NewEnclaveInfo(
			[]byte("productID"),
			[]byte("uniqueID"),
		),
	}
	statusService := statusService{
		Service: svc,
	}

	res, err := statusService.GetStatus(context.Background(), nil)
	require.NoError(t, err)
	require.Equal(t, oracleAcc.GetAddress(), res.OracleAccountAddress)
	require.Equal(t, conf.API.Enabled, res.Api.Enabled)
	require.Equal(t, conf.API.ListenAddr, res.Api.ListenAddr)
	require.Equal(t, conf.GRPC.ListenAddr, res.Grpc.ListenAddr)
	require.Equal(t, svc.EnclaveInfo.ProductID, res.EnclaveInfo.ProductId)
	require.Equal(t, svc.EnclaveInfo.UniqueIDHex(), res.EnclaveInfo.UniqueId)
}
