package event

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/medibloc/panacea-oracle/config"
	"github.com/medibloc/panacea-oracle/ipfs"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/sgx"
)

// Reactor contains all ingredients needed for handling all event types
type Reactor interface {
	GRPCClient() *panacea.GRPCClient
	EnclaveInfo() *sgx.EnclaveInfo
	OracleAcc() *panacea.OracleAccount
	OraclePrivKey() *btcec.PrivateKey
	Config() *config.Config
	QueryClient() panacea.QueryClient
	IPFS() *ipfs.IPFS
	BroadcastTx(txBytes []byte) (int64, string, error)
}

type OracleService interface {
	Reactor
	GetAndStoreOraclePrivKey() error
}
