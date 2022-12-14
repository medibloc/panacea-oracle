package event

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/medibloc/panacea-oracle/config"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/sgx"
)

type RegisterOracleService interface {
	EnclaveInfo() *sgx.EnclaveInfo
	OracleAcc() *panacea.OracleAccount
	OraclePrivKey() *btcec.PrivateKey
	QueryClient() panacea.QueryClient
	Config() *config.Config
	BroadcastTx(txBytes []byte) (int64, string, error)
}

type ApproveOracleRegistrationService interface {
	EnclaveInfo() *sgx.EnclaveInfo
	OracleAcc() *panacea.OracleAccount
	GetAndStoreOraclePrivKey() error
}
