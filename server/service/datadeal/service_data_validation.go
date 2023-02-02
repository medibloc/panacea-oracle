package datadeal

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/gogo/protobuf/proto"
	datadealtypes "github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/medibloc/panacea-oracle/crypto"
	datadeal "github.com/medibloc/panacea-oracle/pb/datadeal/v0"
	"github.com/medibloc/panacea-oracle/server/rpc/interceptor/auth"
	"github.com/medibloc/panacea-oracle/server/service/key"
	"github.com/medibloc/panacea-oracle/validation"
	log "github.com/sirupsen/logrus"
)

func (s *dataDealServiceServer) ValidateData(ctx context.Context, req *datadeal.ValidateDataRequest) (*datadeal.ValidateDataResponse, error) {
	queryClient := s.QueryClient()
	oraclePrivKey := s.OraclePrivKey()
	dealID := req.DealId

	log.Infof("ValidateDataRequest: %v", req)
	if err := req.ValidateBasic(); err != nil {
		log.Errorf("invalid request body: %s", err.Error())
		return nil, err
	}

	requesterAddress, err := auth.GetRequestAddress(ctx)
	if err != nil {
		log.Errorf("failed to get request address. %v", err.Error())
		return nil, err
	}

	if requesterAddress != req.ProviderAddress {
		err := fmt.Errorf("data provider and token issuer do not matched. provider: %s, jwt issuer: %s", req.ProviderAddress, requesterAddress)
		log.Error(err)
		return nil, err
	}

	deal, err := queryClient.GetDeal(ctx, dealID)
	if err != nil {
		log.Errorf("failed to get deal(%d): %s", dealID, err.Error())
		return nil, fmt.Errorf("failed to get deal. %w", err)
	}

	if deal.Status != datadealtypes.DEAL_STATUS_ACTIVE {
		log.Errorf("cannot provide data to INACTIVE/COMPLETED deal")
		return nil, fmt.Errorf("cannot provide data to INACTIVE/COMPLETED deal")
	}

	// Decrypt data
	encryptedData := req.EncryptedData

	providerAcc, err := queryClient.GetAccount(ctx, req.ProviderAddress)
	if err != nil {
		log.Errorf("failed to get provider's account: %s", err.Error())
		return nil, fmt.Errorf("failed to get provider's account")
	}

	if providerAcc.GetPubKey() == nil {
		log.Errorf("failed to get public key of provider's account: %s", err.Error())
		return nil, fmt.Errorf("failed to get public key of provider's account")
	}

	providerPubKeyBytes := providerAcc.GetPubKey().Bytes()

	providerPubKey, err := btcec.ParsePubKey(providerPubKeyBytes, btcec.S256())
	if err != nil {
		log.Errorf("failed to parse provider's public key: %s", err.Error())
		return nil, fmt.Errorf("failed to parse provider's public key")
	}

	decryptSharedKey := crypto.DeriveSharedKey(oraclePrivKey, providerPubKey, crypto.KDFSHA256)

	decryptedData, err := crypto.Decrypt(decryptSharedKey, nil, encryptedData)
	if err != nil {
		log.Errorf("failed to decrypt data: %s", err.Error())
		return nil, fmt.Errorf("failed to decrypt data")
	}

	// Validate data
	dataHash := sha256.Sum256(decryptedData)

	if bytes.Equal(req.DataHash, dataHash[:]) {
		log.Errorf("data hash mismatch")
		return nil, fmt.Errorf("data hash mismatch")
	}

	if err := validation.ValidateJSONSchemata(decryptedData, deal.DataSchema); err != nil {
		log.Errorf("failed to validate data: %s", err.Error())
		return nil, fmt.Errorf("failed to validate data")
	}

	// Re-encrypt data using a combined key
	combinedKey := key.GetCombinedKey(oraclePrivKey.Serialize(), dealID, dataHash[:])
	reEncryptedData, err := crypto.Encrypt(combinedKey[:], nil, decryptedData)
	if err != nil {
		log.Errorf("failed to re-encrypt data with the combined key: %s", err.Error())
		return nil, fmt.Errorf("failed to re-encrypt data with the combined key")
	}

	// Put data into IPFS
	cid, err := s.IPFS().Add(reEncryptedData)
	if err != nil {
		log.Errorf("failed to store data to IPFS: %s", err.Error())
		return nil, fmt.Errorf("failed to store data to IPFS")
	}

	// Issue a certificate to the client
	unsignedDataCert := &datadeal.UnsignedCertificate{
		Cid:             cid,
		UniqueId:        s.EnclaveInfo().UniqueIDHex(),
		OracleAddress:   s.OracleAcc().GetAddress(),
		DealId:          dealID,
		ProviderAddress: req.ProviderAddress,
		DataHash:        req.DataHash,
	}
	key, _ := btcec.PrivKeyFromBytes(btcec.S256(), oraclePrivKey.Serialize())

	marshaledDataCert, err := proto.Marshal(unsignedDataCert)
	if err != nil {
		log.Errorf("failed to marshal data certificate: %s", err.Error())
		return nil, fmt.Errorf("failed to marshal data certificate")
	}

	sig, err := key.Sign(marshaledDataCert)
	if err != nil {
		log.Errorf("failed to create signature of data certificate: %s", err.Error())
		return nil, fmt.Errorf("failed to create signature of data certificate")
	}

	certificate := &datadeal.Certificate{
		UnsignedCertificate: unsignedDataCert,
		Signature:           sig.Serialize(),
	}

	return &datadeal.ValidateDataResponse{
		Certificate: certificate,
	}, nil
}