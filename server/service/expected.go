package service

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/medibloc/panacea-oracle/config"
	"github.com/medibloc/panacea-oracle/ipfs"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/sgx"
)

// Service is 'service/service.go'
type Service interface {
	GetEnclaveInfo() *sgx.EnclaveInfo
	GetOracleAcc() *panacea.OracleAccount
	GetOraclePrivKey() *btcec.PrivateKey
	GetConfig() *config.Config
	GetQueryClient() panacea.QueryClient
	GetIPFS() ipfs.IPFS
}
