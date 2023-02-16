package crypto_test

import (
	"crypto/rand"
	"io"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/stretchr/testify/require"
)

func TestEncryptData(t *testing.T) {
	privKey1, err := btcec.NewPrivateKey(btcec.S256())
	require.NoError(t, err)
	privKey2, err := btcec.NewPrivateKey(btcec.S256())
	require.NoError(t, err)

	data := []byte("This is temporary data")

	shareKey1 := crypto.DeriveSharedKey(privKey1, privKey2.PubKey(), crypto.KDFSHA256)
	shareKey2 := crypto.DeriveSharedKey(privKey2, privKey1.PubKey(), crypto.KDFSHA256)

	nonce := make([]byte, 12)
	_, err = io.ReadFull(rand.Reader, nonce)
	require.NoError(t, err)

	encryptedData, err := crypto.Encrypt(shareKey1, nonce, data)
	require.NoError(t, err)

	decryptedData, err := crypto.Decrypt(shareKey2, nonce, encryptedData)
	require.NoError(t, err)

	require.Equal(t, decryptedData, data)
}
