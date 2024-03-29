package crypto

import (
	"crypto/sha256"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/go-bip39"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

// NewPrivKey generates a random secp256k1 private key
func NewPrivKey() (*btcec.PrivateKey, error) {
	return btcec.NewPrivateKey(btcec.S256())
}

func PrivKeyFromBytes(privKeyBz []byte) (*btcec.PrivateKey, *btcec.PublicKey) {
	return btcec.PrivKeyFromBytes(btcec.S256(), privKeyBz)
}

func GeneratePrivateKeyFromMnemonic(mnemonic string, coinType, accNum, index uint32) (secp256k1.PrivKey, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic")
	}

	hdPath := hd.NewFundraiserParams(accNum, coinType, index).String()
	master, ch := hd.ComputeMastersFromSeed(bip39.NewSeed(mnemonic, ""))

	return hd.DerivePrivateKeyForPath(master, ch, hdPath)
}

// DeriveSharedKey derives a shared key (which can be used for asymmetric encryption)
// using a specified KDF (Key Derivation Function)
// from a shared secret generated by Diffie-Hellman key exchange (ECDH).
func DeriveSharedKey(priv *btcec.PrivateKey, pub *btcec.PublicKey, kdf func([]byte) []byte) []byte {
	sharedSecret := btcec.GenerateSharedSecret(priv, pub)
	return kdf(sharedSecret)
}

// KDFSHA256 is a key derivation function which uses SHA256.
func KDFSHA256(in []byte) []byte {
	out := sha256.Sum256(in)
	return out[:]
}

func NewMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", err
	}
	return bip39.NewMnemonic(entropy)
}
