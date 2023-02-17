package mocks

import (
	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-oracle/config"
	"github.com/medibloc/panacea-oracle/event"
	"github.com/medibloc/panacea-oracle/ipfs"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/service"
	"github.com/medibloc/panacea-oracle/sgx"
)

// MockService is a very simple mock structure.
// It is implemented to return the value as it is declared in this mock structure.
type MockService struct {
	GrpcClient  *MockGrpcClient
	QueryClient *MockQueryClient
	Sgx         *MockSGX
	Ipfs        *MockIPFS

	Config *config.Config

	EnclaveInfo *sgx.EnclaveInfo

	OracleAccount *panacea.OracleAccount
	OraclePrivKey *btcec.PrivateKey
	NodePrivKey   *btcec.PrivateKey

	BroadcastCode        int64
	BroadcastDescription string
	BroadcastError       error

	broadcastMsgs []sdk.Msg
}

var _ service.Service = &MockService{}

func (m *MockService) GetGRPCClient() panacea.GRPCClient {
	return m.GrpcClient
}

func (m *MockService) GetEnclaveInfo() *sgx.EnclaveInfo {
	return m.EnclaveInfo
}

func (m *MockService) GetSgx() sgx.Sgx {
	return m.Sgx
}

func (m *MockService) GetOracleAcc() *panacea.OracleAccount {
	return m.OracleAccount
}

func (m *MockService) GetOraclePrivKey() *btcec.PrivateKey {
	return m.OraclePrivKey
}

func (m *MockService) GetConfig() *config.Config {
	return m.Config
}

func (m *MockService) GetQueryClient() panacea.QueryClient {
	return m.QueryClient
}

func (m *MockService) GetIPFS() ipfs.IPFS {
	return m.Ipfs
}

func (m *MockService) BroadcastTx(msg ...sdk.Msg) (int64, string, error) {
	m.broadcastMsgs = append(m.broadcastMsgs, msg...)
	return m.BroadcastCode, m.BroadcastDescription, m.BroadcastError
}

func (m *MockService) GetBroadCastTxMsgs() []sdk.Msg {
	return m.broadcastMsgs
}

func (m *MockService) StartSubscriptions(event ...event.Event) error {
	return nil
}

func (m *MockService) Close() error {
	return nil
}
