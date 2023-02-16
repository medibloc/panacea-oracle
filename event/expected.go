package event

import (
	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-oracle/config"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/sgx"
)

// Service is 'service/service.go'
type Service interface {
	GetEnclaveInfo() *sgx.EnclaveInfo
	GetSgx() sgx.Sgx
	GetOracleAcc() *panacea.OracleAccount
	GetOraclePrivKey() *btcec.PrivateKey
	GetQueryClient() panacea.QueryClient
	GetConfig() *config.Config
	BroadcastTx(...sdk.Msg) (int64, string, error)
}
