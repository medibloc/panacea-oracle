package oracle

import (
	"fmt"

	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/medibloc/panacea-oracle/event"
	"github.com/medibloc/panacea-oracle/sgx"
	"github.com/tendermint/tendermint/libs/os"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var _ event.Event = (*ApproveOracleRegistrationEvent)(nil)

type ApproveOracleRegistrationEvent struct {
	reactor  event.Reactor
	doneChan chan error
}

func NewApproveOracleRegistrationEvent(s event.Reactor, doneChan chan error) ApproveOracleRegistrationEvent {
	return ApproveOracleRegistrationEvent{s, doneChan}
}

func (e ApproveOracleRegistrationEvent) Name() string {
	return "ApproveOracleRegistrationEvent"
}

func (e ApproveOracleRegistrationEvent) GetEventQuery() string {
	return fmt.Sprintf("message.action = 'ApproveOracleRegistration' and %s.%s = '%s' and %s.%s = '%s'",
		oracletypes.EventTypeApproveOracleRegistration,
		oracletypes.AttributeKeyOracleAddress,
		e.reactor.OracleAcc().GetAddress(),
		oracletypes.EventTypeApproveOracleRegistration,
		oracletypes.AttributeKeyUniqueID,
		e.reactor.EnclaveInfo().UniqueIDHex(),
	)
}

func (e ApproveOracleRegistrationEvent) EventHandler(event ctypes.ResultEvent) error {
	e.doneChan <- e.getAndStoreOraclePrivKey()
	return nil
}

func (e ApproveOracleRegistrationEvent) getAndStoreOraclePrivKey() error {
	oraclePrivKeyBz, err := e.getOraclePrivKey()
	if err != nil {
		return err
	}

	if err := sgx.SealToFile(oraclePrivKeyBz, e.reactor.Config().AbsOraclePrivKeyPath()); err != nil {
		return fmt.Errorf("failed to seal oraclePrivKey to file. %w", err)
	}

	return nil
}

func (e ApproveOracleRegistrationEvent) getOraclePrivKey() ([]byte, error) {
	oraclePrivKeyPath := e.reactor.Config().AbsOraclePrivKeyPath()
	if os.FileExists(oraclePrivKeyPath) {
		return nil, fmt.Errorf("the oracle private key already exists")
	}

	shareKeyBz, err := e.deriveSharedKey()
	if err != nil {
		return nil, err
	}

	encryptedOraclePrivKeyBz, err := e.getEncryptedOraclePrivKey()
	if err != nil {
		return nil, err
	}

	return crypto.Decrypt(shareKeyBz, nil, encryptedOraclePrivKeyBz)
}

func (e ApproveOracleRegistrationEvent) deriveSharedKey() ([]byte, error) {
	nodePrivKeyPath := e.reactor.Config().AbsNodePrivKeyPath()
	if !os.FileExists(nodePrivKeyPath) {
		return nil, fmt.Errorf("the node private key is not exists")
	}
	nodePrivKeyBz, err := sgx.UnsealFromFile(nodePrivKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to unseal nodePrivKey from file.%w", err)
	}
	nodePrivKey, _ := crypto.PrivKeyFromBytes(nodePrivKeyBz)

	oraclePublicKey, err := e.reactor.QueryClient().GetOracleParamsPublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get oraclePublicKey. %w", err)
	}

	shareKeyBz := crypto.DeriveSharedKey(nodePrivKey, oraclePublicKey, crypto.KDFSHA256)
	return shareKeyBz, nil
}

func (e ApproveOracleRegistrationEvent) getEncryptedOraclePrivKey() ([]byte, error) {
	uniqueID := e.reactor.EnclaveInfo().UniqueIDHex()
	oracleAddress := e.reactor.OracleAcc().GetAddress()
	oracleRegistration, err := e.reactor.QueryClient().GetOracleRegistration(uniqueID, oracleAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get oracleRegistration. %w", err)
	}

	return oracleRegistration.EncryptedOraclePrivKey, nil
}
