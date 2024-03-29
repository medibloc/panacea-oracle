package mocks

import (
	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-oracle/config"
	"github.com/medibloc/panacea-oracle/consumer_service"
	"github.com/medibloc/panacea-oracle/event"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/service"
	"github.com/medibloc/panacea-oracle/sgx"
)

var _ service.Service = &MockService{}

// MockService is a very simple mock structure.
// It is implemented to return the value as it is declared in this mock structure.
type MockService struct {
	grpcClient      *MockGrpcClient
	queryClient     *MockQueryClient
	consumerService *MockConsumerService
	sgx             *MockSGX

	config *config.Config

	enclaveInfo *sgx.EnclaveInfo

	oracleAccount *panacea.OracleAccount
	oraclePrivKey *btcec.PrivateKey
	nodePrivKey   *btcec.PrivateKey

	broadcastTxResponse *MockBroadcastTxResponse

	broadcastMsgs []sdk.Msg
}

func NewMockService(
	grpcClient *MockGrpcClient,
	queryClient *MockQueryClient,
	consumerService *MockConsumerService,
	sgx *MockSGX,
	conf *config.Config,
	enclaveInfo *sgx.EnclaveInfo,
	oracleAccount *panacea.OracleAccount,
	oraclePrivKey *btcec.PrivateKey,
	nodePrivKey *btcec.PrivateKey,
) *MockService {
	return &MockService{
		grpcClient:      grpcClient,
		queryClient:     queryClient,
		consumerService: consumerService,
		sgx:             sgx,
		config:          conf,
		enclaveInfo:     enclaveInfo,
		oracleAccount:   oracleAccount,
		oraclePrivKey:   oraclePrivKey,
		nodePrivKey:     nodePrivKey,
		broadcastMsgs:   make([]sdk.Msg, 0),
	}
}

// SetBroadcastTxResponse sets the result after running BroadcastTx
func (m *MockService) SetBroadcastTxResponse(
	code int64,
	description string,
	err error,
) {
	m.broadcastTxResponse = &MockBroadcastTxResponse{
		code:        code,
		description: description,
		error:       err,
	}
}

func (m *MockService) GRPCClient() panacea.GRPCClient {
	return m.grpcClient
}

func (m *MockService) EnclaveInfo() *sgx.EnclaveInfo {
	return m.enclaveInfo
}

func (m *MockService) SGX() sgx.Sgx {
	return m.sgx
}

func (m *MockService) OracleAcc() *panacea.OracleAccount {
	return m.oracleAccount
}

func (m *MockService) OraclePrivKey() *btcec.PrivateKey {
	return m.oraclePrivKey
}

func (m *MockService) Config() *config.Config {
	return m.config
}

func (m *MockService) QueryClient() panacea.QueryClient {
	return m.queryClient
}

func (m *MockService) ConsumerService() consumer_service.FileStorage {
	return m.consumerService
}

func (m *MockService) BroadcastTx(msg ...sdk.Msg) (int64, string, error) {
	m.broadcastMsgs = append(m.broadcastMsgs, msg...)
	tx := m.broadcastTxResponse
	return tx.code, tx.description, tx.error
}

// BroadCastTxMsgs returns the Tx messages for which it ran BroadcastTx
func (m *MockService) BroadCastTxMsgs() []sdk.Msg {
	return m.broadcastMsgs
}

func (m *MockService) StartSubscriptions(event ...event.Event) error {
	return nil
}

func (m *MockService) Close() error {
	return nil
}
