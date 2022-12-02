package panacea

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/medibloc/panacea-oracle/crypto"
	log "github.com/sirupsen/logrus"
)

type OracleAccount struct {
	privKey cryptotypes.PrivKey
	pubKey  cryptotypes.PubKey
}

// NewOracleAccount returns an oracle account from mnemonic, account number, and index
func NewOracleAccount(mnemonic string, accNum, index uint32) (*OracleAccount, error) {
	if len(mnemonic) == 0 {
		return &OracleAccount{}, fmt.Errorf("mnemonic is empty")
	}

	key, err := crypto.GeneratePrivateKeyFromMnemonic(mnemonic, CoinType, accNum, index)
	if err != nil {
		return &OracleAccount{}, err
	}

	pk := &secp256k1.PrivKey{
		Key: key.Bytes(),
	}

	return &OracleAccount{
		privKey: pk,
		pubKey:  pk.PubKey(),
	}, nil
}

func (oa OracleAccount) GetAddress() string {
	address, err := bech32.ConvertAndEncode(prefix, oa.pubKey.Address().Bytes())
	if err != nil {
		log.Panic(err)
	}

	return address
}

func (oa OracleAccount) AccAddressFromBech32() sdk.AccAddress {
	return oa.pubKey.Bytes()
}

func (oa OracleAccount) GetPrivKey() cryptotypes.PrivKey {
	return oa.privKey
}

func (oa OracleAccount) GetPubKey() cryptotypes.PubKey {
	return oa.pubKey
}

func GetAccAddressFromBech32(address string) (addr sdk.AccAddress, err error) {
	return sdk.GetFromBech32(address, prefix)
}
