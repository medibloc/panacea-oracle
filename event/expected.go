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
	EnclaveInfo() *sgx.EnclaveInfo
	OracleAcc() *panacea.OracleAccount
	OraclePrivKey() *btcec.PrivateKey
	QueryClient() panacea.QueryClient
	Config() *config.Config
	BroadcastTx(...sdk.Msg) (int64, string, error)
}

// OracleService is 'service/oracle/service.go'
type OracleService interface {
	EnclaveInfo() *sgx.EnclaveInfo
	OracleAcc() *panacea.OracleAccount
	GetAndStoreOraclePrivKey() error
}
