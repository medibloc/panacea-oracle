package service

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/medibloc/panacea-oracle/ipfs"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/sgx"
)

type Reactor interface {
	EnclaveInfo() *sgx.EnclaveInfo
	OracleAcc() *panacea.OracleAccount
	OraclePrivKey() *btcec.PrivateKey
	QueryClient() panacea.QueryClient
	IPFS() *ipfs.IPFS
}
