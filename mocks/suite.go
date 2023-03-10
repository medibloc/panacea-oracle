package mocks

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec"
	"github.com/medibloc/panacea-oracle/config"
	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/sgx"
	"github.com/stretchr/testify/suite"
)

type MockTestSuite struct {
	suite.Suite

	UniqueID string

	QueryClient     *MockQueryClient
	GrpcClient      *MockGrpcClient
	ConsumerService *MockConsumerService
	Svc             *MockService
	SGX             *MockSGX
	Config          *config.Config
	EnclaveInfo     *sgx.EnclaveInfo

	OracleAcc     *panacea.OracleAccount
	OraclePrivKey *btcec.PrivateKey
	OraclePubKey  *btcec.PublicKey
	NodePrivKey   *btcec.PrivateKey
}

func (suite *MockTestSuite) Initialize() {
	mnemonic, _ := crypto.NewMnemonic()
	uniqueID := []byte("uniqueID")

	suite.UniqueID = hex.EncodeToString(uniqueID)
	suite.QueryClient = &MockQueryClient{}
	suite.GrpcClient = &MockGrpcClient{}
	suite.SGX = &MockSGX{}
	suite.Config = config.DefaultConfig()
	suite.EnclaveInfo = sgx.NewEnclaveInfo(nil, uniqueID)
	suite.OracleAcc, _ = panacea.NewOracleAccount(mnemonic, 0, 0)
	suite.OraclePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	suite.OraclePubKey = suite.OraclePrivKey.PubKey()
	suite.NodePrivKey, _ = btcec.NewPrivateKey(btcec.S256())

	suite.ConsumerService = &MockConsumerService{suite.OraclePrivKey}

	suite.Svc = NewMockService(
		suite.GrpcClient,
		suite.QueryClient,
		suite.ConsumerService,
		suite.SGX,
		suite.Config,
		suite.EnclaveInfo,
		suite.OracleAcc,
		suite.OraclePrivKey,
		suite.NodePrivKey,
	)

}
