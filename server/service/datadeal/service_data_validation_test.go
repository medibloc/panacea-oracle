package datadeal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	datadealtypes "github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/medibloc/panacea-oracle/panacea"
	datadeal "github.com/medibloc/panacea-oracle/pb/datadeal/v0"
	"github.com/medibloc/panacea-oracle/server/rpc/interceptor/auth"
	"github.com/medibloc/panacea-oracle/server/service/key"
	"github.com/medibloc/panacea-oracle/sgx"
	"github.com/medibloc/panacea-oracle/types/test_utils/mocks"
	"github.com/stretchr/testify/suite"
)

type dataDealServiceServerTestSuite struct {
	suite.Suite

	svc *mocks.MockService

	productID []byte
	uniqueID  []byte

	oracleAcc     *panacea.OracleAccount
	oraclePrivKey *btcec.PrivateKey

	deal *datadealtypes.Deal

	providerAccPrivKey secp256k1.PrivKey
	providerAccPubKey  cryptotypes.PubKey
	providerAcc        authtypes.AccountI
}

func TestDataDealServiceServer(t *testing.T) {
	suite.Run(t, &dataDealServiceServerTestSuite{})
}

func (suite *dataDealServiceServerTestSuite) BeforeTest(_, _ string) {
	suite.productID = []byte("productID")
	suite.uniqueID = []byte("uniqueID")

	mnemonic, _ := crypto.NewMnemonic()
	suite.oracleAcc, _ = panacea.NewOracleAccount(mnemonic, 0, 0)
	suite.oraclePrivKey, _ = btcec.NewPrivateKey(btcec.S256())

	suite.deal = &datadealtypes.Deal{
		Id:         1,
		DataSchema: []string{"https://json.schemastore.org/github-issue-forms.json"},
		Status:     datadealtypes.DEAL_STATUS_ACTIVE,
	}
	suite.providerAccPrivKey = *secp256k1.GenPrivKey()
	suite.providerAccPubKey = suite.providerAccPrivKey.PubKey()
	suite.providerAcc = mocks.NewMockAccount(suite.providerAccPubKey)
	suite.svc = &mocks.MockService{
		QueryClient: &mocks.MockQueryClient{
			Account: suite.providerAcc,
			Deal:    suite.deal,
		},
		Ipfs:          &mocks.MockIPFS{},
		EnclaveInfo:   sgx.NewEnclaveInfo(suite.productID, suite.uniqueID),
		OracleAccount: suite.oracleAcc,
		OraclePrivKey: suite.oraclePrivKey,
	}
}

func (suite *dataDealServiceServerTestSuite) AfterTest(_, _ string) {
	mocks.RemoveMockIPFSData()
}

func (suite *dataDealServiceServerTestSuite) TestValidateDataSuccess() {
	// provide data
	jsonDataBz := []byte(
		`
		{
			"name": "name",
			"description": "description",
			"body": [{ "type": "markdown", "attributes": { "value": "val1" } }]
		}
		`)

	// encrypted provider data with provider private key and oracle public key
	providerPrivKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), suite.providerAccPrivKey.Bytes())

	sharedKey := crypto.DeriveSharedKey(
		providerPrivKey,
		suite.oraclePrivKey.PubKey(),
		crypto.KDFSHA256,
	)

	encryptedData, err := crypto.Encrypt(sharedKey, nil, jsonDataBz)
	suite.Require().NoError(err)

	dataHash := sha256.Sum256(jsonDataBz)

	req := &datadeal.ValidateDataRequest{
		DealId:          1,
		ProviderAddress: panacea.GetAddressFromPrivateKey(suite.providerAccPrivKey),
		EncryptedData:   encryptedData,
		DataHash:        dataHash[:],
	}

	// add authentication in header
	ctx := context.Background()
	ctx = context.WithValue(ctx, auth.ContextKeyAuthenticatedAccountAddress{}, req.ProviderAddress)

	// request validation for provider data
	server := dataDealServiceServer{Service: suite.svc}
	res, err := server.ValidateData(ctx, req)
	suite.Require().NoError(err)

	// compare certificate
	unsignedCertificate := res.Certificate.UnsignedCertificate
	suite.Require().Equal(hex.EncodeToString(suite.uniqueID), unsignedCertificate.UniqueId)
	suite.Require().Equal(suite.oracleAcc.GetAddress(), unsignedCertificate.OracleAddress)
	suite.Require().Equal(req.DealId, unsignedCertificate.DealId)
	suite.Require().Equal(req.ProviderAddress, unsignedCertificate.ProviderAddress)
	suite.Require().Equal(hex.EncodeToString(req.DataHash), unsignedCertificate.DataHash)
	suite.Require().NotNil(res.Certificate.Signature)

	// verify certificate
	marshal, err := unsignedCertificate.Marshal()
	suite.Require().NoError(err)
	signature, err := btcec.ParseSignature(res.Certificate.Signature, btcec.S256())
	suite.Require().NoError(err)
	suite.Require().True(signature.Verify(marshal, suite.oraclePrivKey.PubKey()))

	// decrypt re-encrypted provider's data
	reEncryptedData, err := suite.svc.GetIPFS().Get(unsignedCertificate.Cid)
	suite.Require().NoError(err)
	combinedKey := key.GetSecretKey(suite.oraclePrivKey.Serialize(), req.DealId, dataHash[:])
	decryptedData, err := crypto.Decrypt(combinedKey[:], nil, reEncryptedData)
	suite.Require().NoError(err)
	suite.Require().Equal(jsonDataBz, decryptedData)
}

func (suite *dataDealServiceServerTestSuite) TestValidateDataInvalidRequest() {
	req := &datadeal.ValidateDataRequest{
		DealId:          1,
		ProviderAddress: "invalid_provider_address",
		EncryptedData:   nil,
		DataHash:        nil,
	}

	ctx := context.Background()

	// request validation for provider data
	server := dataDealServiceServer{Service: suite.svc}
	res, err := server.ValidateData(ctx, req)
	suite.Require().Nil(res)
	suite.Require().ErrorContains(err, "invalid provider address:")

	req.ProviderAddress = panacea.GetAddressFromPrivateKey(suite.providerAccPrivKey)
	res, err = server.ValidateData(ctx, req)
	suite.Require().Nil(res)
	suite.Require().ErrorContains(err, "encrypted data is empty in request")

	req.EncryptedData = []byte("encryptedData") // only check length
	res, err = server.ValidateData(ctx, req)
	suite.Require().Nil(res)
	suite.Require().ErrorContains(err, "data hash is empty in request")

	req.DataHash = []byte("dataHash")
	res, err = server.ValidateData(ctx, req)
	suite.Require().Nil(res)
	suite.Require().ErrorContains(err, "failed to get request address")

	ctx = context.WithValue(ctx, auth.ContextKeyAuthenticatedAccountAddress{}, "invalid provider address")
	res, err = server.ValidateData(ctx, req)
	suite.Require().Nil(res)
	suite.Require().ErrorContains(err, "data provider and token issuer do not matched")
}

func (suite *dataDealServiceServerTestSuite) TestValidateDataDealStatusIsNotActive() {
	// set deal
	suite.deal.Status = datadealtypes.DEAL_STATUS_INACTIVE

	req := &datadeal.ValidateDataRequest{
		DealId:          1,
		ProviderAddress: panacea.GetAddressFromPrivateKey(suite.providerAccPrivKey),
		EncryptedData:   []byte("encryptedData"),
		DataHash:        []byte("dataHash"),
	}

	// add authentication in header
	ctx := context.Background()
	ctx = context.WithValue(ctx, auth.ContextKeyAuthenticatedAccountAddress{}, req.ProviderAddress)

	// request validation for provider data
	server := dataDealServiceServer{Service: suite.svc}
	res, err := server.ValidateData(ctx, req)
	suite.Require().Nil(res)
	suite.Require().ErrorContains(err, "cannot provide data to INACTIVE/COMPLETED deal")

	suite.deal.Status = datadealtypes.DEAL_STATUS_COMPLETED
	res, err = server.ValidateData(ctx, req)
	suite.Require().Nil(res)
	suite.Require().ErrorContains(err, "cannot provide data to INACTIVE/COMPLETED deal")
}

func (suite *dataDealServiceServerTestSuite) TestValidateDataNotFoundProviderPublicKey() {
	req := &datadeal.ValidateDataRequest{
		DealId:          1,
		ProviderAddress: panacea.GetAddressFromPrivateKey(suite.providerAccPrivKey),
		EncryptedData:   []byte("encryptedData"),
		DataHash:        []byte("dataHash"),
	}

	// add authentication in header
	ctx := context.Background()
	ctx = context.WithValue(ctx, auth.ContextKeyAuthenticatedAccountAddress{}, req.ProviderAddress)

	// set provider public key to nil
	suite.svc.QueryClient.Account = authtypes.NewBaseAccount(
		sdk.AccAddress(suite.providerAccPubKey.Address()),
		nil,
		1,
		1,
	)

	// request validation for provider data
	server := dataDealServiceServer{Service: suite.svc}
	res, err := server.ValidateData(ctx, req)
	suite.Require().Nil(res)
	suite.Require().ErrorContains(err, "failed to get public key of provider's account")
}

func (suite *dataDealServiceServerTestSuite) TestValidateDataInvalidProviderEncryptedData() {
	req := &datadeal.ValidateDataRequest{
		DealId:          1,
		ProviderAddress: panacea.GetAddressFromPrivateKey(suite.providerAccPrivKey),
		EncryptedData:   []byte("encryptedData"),
		DataHash:        []byte("dataHash"),
	}

	// add authentication in header
	ctx := context.Background()
	ctx = context.WithValue(ctx, auth.ContextKeyAuthenticatedAccountAddress{}, req.ProviderAddress)

	// request validation for provider data
	server := dataDealServiceServer{Service: suite.svc}
	res, err := server.ValidateData(ctx, req)
	suite.Require().Nil(res)
	suite.Require().ErrorContains(err, "failed to decrypt data")
}

func (suite *dataDealServiceServerTestSuite) TestValidateDataNotMatchedDataHash() {
	// provide data
	jsonDataBz := []byte(
		`
		{
			"name": "name",
			"description": "description",
			"body": [{ "type": "markdown", "attributes": { "value": "val1" } }]
		}
		`)

	// encrypted provider data with provider private key and oracle public key
	providerPrivKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), suite.providerAccPrivKey.Bytes())

	sharedKey := crypto.DeriveSharedKey(
		providerPrivKey,
		suite.oraclePrivKey.PubKey(),
		crypto.KDFSHA256,
	)

	encryptedData, err := crypto.Encrypt(sharedKey, nil, jsonDataBz)
	suite.Require().NoError(err)

	req := &datadeal.ValidateDataRequest{
		DealId:          1,
		ProviderAddress: panacea.GetAddressFromPrivateKey(suite.providerAccPrivKey),
		EncryptedData:   encryptedData,
		DataHash:        []byte("invalid data hash"),
	}

	// add authentication in header
	ctx := context.Background()
	ctx = context.WithValue(ctx, auth.ContextKeyAuthenticatedAccountAddress{}, req.ProviderAddress)

	// request validation for provider data
	server := dataDealServiceServer{Service: suite.svc}
	res, err := server.ValidateData(ctx, req)
	suite.Require().Nil(res)
	suite.Require().ErrorContains(err, "data hash mismatch")
}

func (suite *dataDealServiceServerTestSuite) TestValidateDataInvalidJSONSchema() {
	// provide data
	jsonDataBz := []byte(
		`
		{
			"invalid_key_name": "name",
			"invalid_key_description": "description",
			"invalid_key_body": [{ "type": "markdown", "attributes": { "value": "val1" } }]
		}
		`)

	// encrypted provider data with provider private key and oracle public key
	providerPrivKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), suite.providerAccPrivKey.Bytes())

	sharedKey := crypto.DeriveSharedKey(
		providerPrivKey,
		suite.oraclePrivKey.PubKey(),
		crypto.KDFSHA256,
	)

	encryptedData, err := crypto.Encrypt(sharedKey, nil, jsonDataBz)
	suite.Require().NoError(err)

	dataHash := sha256.Sum256(jsonDataBz)

	req := &datadeal.ValidateDataRequest{
		DealId:          1,
		ProviderAddress: panacea.GetAddressFromPrivateKey(suite.providerAccPrivKey),
		EncryptedData:   encryptedData,
		DataHash:        dataHash[:],
	}

	// add authentication in header
	ctx := context.Background()
	ctx = context.WithValue(ctx, auth.ContextKeyAuthenticatedAccountAddress{}, req.ProviderAddress)

	// request validation for provider data
	server := dataDealServiceServer{Service: suite.svc}
	res, err := server.ValidateData(ctx, req)
	suite.Require().Nil(res)
	suite.Require().ErrorContains(err, "failed to validate data")
}
