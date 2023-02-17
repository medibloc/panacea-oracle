package mocks

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var _ authtypes.AccountI = &MockAccount{}

type MockAccount struct {
	*authtypes.BaseAccount
}

func NewMockAccount(pubKey cryptotypes.PubKey) *MockAccount {
	return &MockAccount{
		authtypes.NewBaseAccount(
			sdk.AccAddress(pubKey.Address()),
			pubKey,
			1,
			1,
		),
	}
}
