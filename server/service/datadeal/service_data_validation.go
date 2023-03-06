package datadeal

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/gogo/protobuf/proto"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	datadealtypes "github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/medibloc/panacea-oracle/panacea"
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

	if err := validateRequest(req); err != nil {
		log.Debugf("invalid request body: %s", err.Error())
		return nil, err
	}

	requesterAddress, err := auth.GetRequestAddress(ctx)
	if err != nil {
		log.Debugf("failed to get request address. %v", err.Error())
		return nil, fmt.Errorf("failed to get request address. %w", err)
	}

	if requesterAddress != req.ProviderAddress {
		log.Debugf("data provider and token issuer do not matched.  provider: %s, jwt issuer: %s", req.ProviderAddress, requesterAddress)
		return nil, fmt.Errorf("data provider and token issuer do not matched.  provider: %s, jwt issuer: %s", req.ProviderAddress, requesterAddress)
	}

	deal, err := queryClient.GetDeal(ctx, dealID)
	if err != nil {
		log.Debugf("failed to get deal(%d): %s", dealID, err.Error())
		return nil, fmt.Errorf("failed to get deal. %w", err)
	}

	if deal.Status != datadealtypes.DEAL_STATUS_ACTIVE {
		log.Debugf("cannot provide data to INACTIVE/COMPLETED deal")
		return nil, fmt.Errorf("cannot provide data to INACTIVE/COMPLETED deal")
	}

	// Decrypt data
	encryptedData := req.EncryptedData

	providerAcc, err := queryClient.GetAccount(ctx, req.ProviderAddress)
	if err != nil {
		log.Debugf("failed to get provider's account: %v", err)
		return nil, fmt.Errorf("failed to get provider's account: %w", err)
	}

	if providerAcc.GetPubKey() == nil {
		log.Debugf("failed to get public key of provider's account: %s", req.ProviderAddress)
		return nil, fmt.Errorf("failed to get public key of provider's account: %s", req.ProviderAddress)
	}

	providerPubKeyBytes := providerAcc.GetPubKey().Bytes()

	providerPubKey, err := btcec.ParsePubKey(providerPubKeyBytes, btcec.S256())
	if err != nil {
		log.Debugf("failed to parse provider's public key: %v", err)
		return nil, fmt.Errorf("failed to parse provider's public key: %w", err)
	}

	decryptSharedKey := crypto.DeriveSharedKey(oraclePrivKey, providerPubKey, crypto.KDFSHA256)

	decryptedData, err := crypto.Decrypt(decryptSharedKey, nil, encryptedData)
	if err != nil {
		log.Debugf("failed to decrypt data: %s", err.Error())
		return nil, fmt.Errorf("failed to decrypt data")
	}

	// Validate data
	dataHashBz := crypto.KDFSHA256(decryptedData)
	dataHash := hex.EncodeToString(dataHashBz)

	if req.DataHash != dataHash {
		log.Errorf("data hash mismatch")
		return nil, fmt.Errorf("data hash mismatch")
	}

	if err := validation.ValidateJSONSchemata(decryptedData, deal.DataSchema); err != nil {
		log.Debugf("failed to validate data: %s", err.Error())
		return nil, fmt.Errorf("failed to validate data")
	}

	// Re-encrypt data using a combined key
	secretKey := key.GetSecretKey(oraclePrivKey.Serialize(), dealID, dataHashBz)
	reEncryptedData, err := crypto.Encrypt(secretKey, nil, decryptedData)
	if err != nil {
		log.Errorf("failed to re-encrypt data with the combined key: %s", err.Error())
		return nil, fmt.Errorf("failed to re-encrypt data with the combined key")
	}

	// Post reEncryptedData to consumer service
	dataUrl := deal.ConsumerServiceEndpoint + "/v1/data/" + strconv.FormatUint(req.DealId, 10) + "/" + req.DataHash
	token, err := generateJWT(oraclePrivKey, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to generate jwt: %v", err.Error())
	}
	if err := postData(reEncryptedData, dataUrl, token); err != nil {
		return nil, fmt.Errorf("failed to post request: %s", err.Error())
	}

	// Issue a certificate to the client
	unsignedDataCert := &datadealtypes.UnsignedCertificate{
		DataEndpoint:    dataUrl,
		UniqueId:        s.EnclaveInfo().UniqueIDHex(),
		OracleAddress:   s.OracleAcc().GetAddress(),
		DealId:          dealID,
		ProviderAddress: req.ProviderAddress,
		DataHash:        dataHash,
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

	certificate := &datadealtypes.Certificate{
		UnsignedCertificate: unsignedDataCert,
		Signature:           sig.Serialize(),
	}

	return &datadeal.ValidateDataResponse{
		Certificate: certificate,
	}, nil
}

func validateRequest(req *datadeal.ValidateDataRequest) error {
	if _, err := panacea.GetAccAddressFromBech32(req.ProviderAddress); err != nil {
		return fmt.Errorf("invalid provider address: %w", err)
	}

	if len(req.EncryptedData) == 0 {
		return fmt.Errorf("encrypted data is empty in request")
	}

	if len(req.DataHash) == 0 {
		return fmt.Errorf("data hash is empty in request")
	}

	return nil
}

func postData(data []byte, dataUrl string, jwt []byte) error {
	buff := bytes.NewBuffer(data)

	request, err := http.NewRequest("POST", dataUrl, buff)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+string(jwt))

	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}
	return nil
}

func generateJWT(privKey *btcec.PrivateKey, expiration time.Duration) ([]byte, error) {
	now := time.Now().Truncate(time.Second)
	token, err := jwt.NewBuilder().
		IssuedAt(now).
		NotBefore(now).
		Expiration(now.Add(expiration)).
		Build()
	if err != nil {
		return nil, fmt.Errorf("build jwt error: %v", err)
	}
	signedJWT, err := jwt.Sign(token, jwt.WithKey(jwa.ES256K, privKey.ToECDSA()))
	if err != nil {
		return nil, fmt.Errorf("jwt signing error: %v", err)
	}
	return signedJWT, nil
}
