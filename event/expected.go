package event

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/medibloc/panacea-oracle/config"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/sgx"
)

// Reactor contains all ingredients needed for handling this type of event
type Reactor interface {
	GRPCClient() *panacea.GrpcClient
	EnclaveInfo() *sgx.EnclaveInfo
	OracleAcc() *panacea.OracleAccount
	OraclePrivKey() *btcec.PrivateKey
	Config() *config.Config
	QueryClient() *panacea.QueryClient
	//Ipfs() *ipfs.ipfs
	BroadcastTx(txBytes []byte) (int64, string, error)
}
