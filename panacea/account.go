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
		return nil, fmt.Errorf("mnemonic is empty")
	}

	key, err := GetPrivateKeyFromMnemonic(mnemonic, accNum, index)
	if err != nil {
		return nil, err
	}

	return &OracleAccount{
		privKey: &key,
		pubKey:  key.PubKey(),
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
	return oa.pubKey.Address().Bytes()
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

func GetAddressFromMnemonic(mnemonic string, accNum, index uint32) string {
	key, err := GetPrivateKeyFromMnemonic(mnemonic, accNum, index)
	if err != nil {
		log.Panic(err)
	}

	return GetAddressFromPrivateKey(key)
}

func GetAddressFromPrivateKey(key secp256k1.PrivKey) string {
	addr, err := bech32.ConvertAndEncode(prefix, key.PubKey().Address().Bytes())
	if err != nil {
		log.Panic(err)
	}
	return addr
}

func GetPrivateKeyFromMnemonic(mnemonic string, accNum, index uint32) (secp256k1.PrivKey, error) {
	key, err := crypto.GeneratePrivateKeyFromMnemonic(mnemonic, CoinType, accNum, index)
	if err != nil {
		return secp256k1.PrivKey{}, err
	}
	return secp256k1.PrivKey{
		Key: key,
	}, nil
}
