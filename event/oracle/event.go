package oracle

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/tendermint/tendermint/light/provider"
)

func makeMsgApproveOracleRegistration(uniqueID, approverAddr, targetAddr string, oraclePrivKey, nodePubKey []byte) (*oracletypes.MsgApproveOracleRegistration, error) {
	privKey, _ := crypto.PrivKeyFromBytes(oraclePrivKey)
	pubKey, err := btcec.ParsePubKey(nodePubKey, btcec.S256())
	if err != nil {
		return nil, err
	}

	shareKey := crypto.DeriveSharedKey(privKey, pubKey, crypto.KDFSHA256)
	encryptedOraclePrivKey, err := crypto.EncryptWithAES256(shareKey, nil, oraclePrivKey)
	if err != nil {
		return nil, err
	}

	registrationApproval := &oracletypes.ApproveOracleRegistration{
		UniqueId:               uniqueID,
		ApproverOracleAddress:  approverAddr,
		TargetOracleAddress:    targetAddr,
		EncryptedOraclePrivKey: encryptedOraclePrivKey,
	}

	return makeMsgApproveOracleRegistrationWithSignature(registrationApproval, oraclePrivKey)
}

func makeMsgApproveOracleRegistrationWithSignature(approveOracleRegistration *oracletypes.ApproveOracleRegistration, oraclePrivKey []byte) (*oracletypes.MsgApproveOracleRegistration, error) {
	key, _ := crypto.PrivKeyFromBytes(oraclePrivKey)

	marshaledApproveOracleRegistration, err := approveOracleRegistration.Marshal()
	if err != nil {
		return nil, err
	}

	sig, err := key.Sign(marshaledApproveOracleRegistration)
	if err != nil {
		return nil, err
	}

	msgApproveOracleRegistration := &oracletypes.MsgApproveOracleRegistration{
		ApproveOracleRegistration: approveOracleRegistration,
		Signature:                 sig.Serialize(),
	}

	return msgApproveOracleRegistration, nil
}

func verifyTrustedBlockInfo(queryClient *panacea.QueryClient, height int64, blockHash []byte) error {
	block, err := queryClient.GetLightBlock(height)
	if err != nil {
		switch err {
		case provider.ErrLightBlockNotFound, provider.ErrHeightTooHigh:
			return fmt.Errorf("not found light block. %w", err)
		default:
			return err
		}
	}

	if !bytes.Equal(block.Hash().Bytes(), blockHash) {
		return fmt.Errorf("failed to verify trusted block information. height(%v), expected block hash(%s), got block hash(%s)",
			height,
			hex.EncodeToString(block.Hash().Bytes()),
			hex.EncodeToString(blockHash),
		)
	}

	return nil
}
