package oracle

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/gogo/protobuf/proto"
	"github.com/medibloc/panacea-oracle/crypto"
)

func encryptOraclePrivKey(oraclePrivKey, nodePubKey []byte) ([]byte, error) {
	privKey, _ := crypto.PrivKeyFromBytes(oraclePrivKey)
	pubKey, err := btcec.ParsePubKey(nodePubKey, btcec.S256())
	if err != nil {
		return nil, err
	}

	sharedKey := crypto.DeriveSharedKey(privKey, pubKey, crypto.KDFSHA256)
	return crypto.Encrypt(sharedKey, nil, oraclePrivKey)
}

func signApprovalMsg(approvalMsg proto.Marshaler, oraclePrivKey []byte) ([]byte, error) {
	key, _ := crypto.PrivKeyFromBytes(oraclePrivKey)

	marshaledApproveMsg, err := approvalMsg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal approval message: %w", err)
	}

	sig, err := key.Sign(marshaledApproveMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to create signature: %w", err)
	}

	return sig.Serialize(), nil
}
