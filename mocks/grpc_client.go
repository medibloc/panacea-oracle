package mocks

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/medibloc/panacea-oracle/panacea"
)

var _ panacea.GRPCClient = &MockGrpcClient{}

// MockGrpcClient is a very simple mock structure.
// It is implemented to return the value as it is declared in this mock structure.
type MockGrpcClient struct {
	BroadcastResponse *tx.BroadcastTxResponse
	ProtoCodec        *codec.ProtoCodec
	ChainID           string
	Account           *MockAccount
}

func (m MockGrpcClient) Close() error {
	return nil
}

func (m MockGrpcClient) BroadcastTx(txBytes []byte) (*tx.BroadcastTxResponse, error) {
	return m.BroadcastResponse, nil
}

func (m MockGrpcClient) GetCdc() *codec.ProtoCodec {
	return m.ProtoCodec
}

func (m MockGrpcClient) GetChainID() string {
	return m.ChainID
}

func (m MockGrpcClient) GetAccount(address string) (authtypes.AccountI, error) {
	return m.Account, nil
}
