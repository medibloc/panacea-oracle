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
	"github.com/cyberphone/json-canonicalization/go/src/webpki.org/jsoncanonicalizer"
	datadealtypes "github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/medibloc/panacea-oracle/mocks"
	"github.com/medibloc/panacea-oracle/panacea"
	datadeal "github.com/medibloc/panacea-oracle/pb/datadeal/v0"
	"github.com/medibloc/panacea-oracle/server/rpc/interceptor/auth"
	"github.com/medibloc/panacea-oracle/server/service/key"
	"github.com/medibloc/panacea-oracle/validation"
	"github.com/stretchr/testify/suite"
)

// TODO: This test will be changed to VP data validation.
type dataDealServiceServerTestSuite struct {
	mocks.MockTestSuite

	deal *datadealtypes.Deal

	providerAccPrivKey secp256k1.PrivKey
	providerAccPubKey  cryptotypes.PubKey
	providerAcc        authtypes.AccountI
}

func TestDataDealServiceServer(t *testing.T) {
	suite.Run(t, &dataDealServiceServerTestSuite{})
}

func (suite *dataDealServiceServerTestSuite) BeforeTest(_, _ string) {
	suite.Initialize()
	tempDir := suite.T().TempDir()

	suite.deal = &datadealtypes.Deal{
		Id:                      1,
		DataSchema:              []string{"https://json.schemastore.org/github-issue-forms.json"},
		Status:                  datadealtypes.DEAL_STATUS_ACTIVE,
		ConsumerServiceEndpoint: tempDir,
	}
	suite.providerAccPrivKey = *secp256k1.GenPrivKey()
	suite.providerAccPubKey = suite.providerAccPrivKey.PubKey()
	suite.providerAcc = mocks.NewMockAccount(suite.providerAccPubKey)
	suite.QueryClient.Account = suite.providerAcc
	suite.QueryClient.Deal = suite.deal

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
		suite.OraclePrivKey.PubKey(),
		crypto.KDFSHA256,
	)

	encryptedData, err := crypto.Encrypt(sharedKey, nil, jsonDataBz)
	suite.Require().NoError(err)

	jsonData, err := jsoncanonicalizer.Transform(jsonDataBz)
	suite.Require().NoError(err)
	dataHash := sha256.Sum256(jsonData)

	req := &datadeal.ValidateDataRequest{
		DealId:          1,
		ProviderAddress: panacea.GetAddressFromPrivateKey(suite.providerAccPrivKey),
		EncryptedData:   encryptedData,
		DataHash:        hex.EncodeToString(dataHash[:]),
	}

	// add authentication in header
	ctx := context.Background()
	ctx = context.WithValue(ctx, auth.ContextKeyAuthenticatedAccountAddress{}, req.ProviderAddress)

	// request validation for provider data
	server := dataDealServiceServer{Service: suite.Svc, schema: validation.NewJSONSchema()}
	res, err := server.ValidateData(ctx, req)
	suite.Require().NoError(err)

	// compare certificate
	unsignedCertificate := res.Certificate.UnsignedCertificate
	suite.Require().Equal(suite.UniqueID, unsignedCertificate.UniqueId)
	suite.Require().Equal(suite.OracleAcc.GetAddress(), unsignedCertificate.OracleAddress)
	suite.Require().Equal(req.DealId, unsignedCertificate.DealId)
	suite.Require().Equal(req.ProviderAddress, unsignedCertificate.ProviderAddress)
	suite.Require().Equal(req.DataHash, unsignedCertificate.DataHash)
	suite.Require().NotNil(res.Certificate.Signature)

	// verify certificate
	marshal, err := unsignedCertificate.Marshal()
	suite.Require().NoError(err)
	signature, err := btcec.ParseSignature(res.Certificate.Signature, btcec.S256())
	suite.Require().NoError(err)
	suite.Require().True(signature.Verify(marshal, suite.OraclePrivKey.PubKey()))

	// decrypt re-encrypted provider's data
	reEncryptedData, err := suite.ConsumerService.Get(suite.deal.ConsumerServiceEndpoint, unsignedCertificate.DealId, unsignedCertificate.DataHash)
	suite.Require().NoError(err)
	combinedKey := key.GetSecretKey(suite.OraclePrivKey.Serialize(), req.DealId, dataHash[:])
	decryptedData, err := crypto.Decrypt(combinedKey[:], nil, reEncryptedData)
	suite.Require().NoError(err)
	suite.Require().Equal(jsonDataBz, decryptedData)
}

func (suite *dataDealServiceServerTestSuite) TestValidateDataInvalidRequest() {
	req := &datadeal.ValidateDataRequest{
		DealId:          1,
		ProviderAddress: "invalid_provider_address",
		EncryptedData:   nil,
		DataHash:        "",
	}

	ctx := context.Background()

	// request validation for provider data
	server := dataDealServiceServer{Service: suite.Svc, schema: validation.NewJSONSchema()}
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

	req.DataHash = "dataHash"
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
		DataHash:        "dataHash",
	}

	// add authentication in header
	ctx := context.Background()
	ctx = context.WithValue(ctx, auth.ContextKeyAuthenticatedAccountAddress{}, req.ProviderAddress)

	// request validation for provider data
	server := dataDealServiceServer{Service: suite.Svc, schema: validation.NewJSONSchema()}
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
		DataHash:        "dataHash",
	}

	// add authentication in header
	ctx := context.Background()
	ctx = context.WithValue(ctx, auth.ContextKeyAuthenticatedAccountAddress{}, req.ProviderAddress)

	// set provider public key to nil
	suite.QueryClient.Account = authtypes.NewBaseAccount(
		sdk.AccAddress(suite.providerAccPubKey.Address()),
		nil,
		1,
		1,
	)

	// request validation for provider data
	server := dataDealServiceServer{Service: suite.Svc, schema: validation.NewJSONSchema()}
	res, err := server.ValidateData(ctx, req)
	suite.Require().Nil(res)
	suite.Require().ErrorContains(err, "failed to get public key of provider's account")
}

func (suite *dataDealServiceServerTestSuite) TestValidateDataInvalidProviderEncryptedData() {
	req := &datadeal.ValidateDataRequest{
		DealId:          1,
		ProviderAddress: panacea.GetAddressFromPrivateKey(suite.providerAccPrivKey),
		EncryptedData:   []byte("encryptedData"),
		DataHash:        "dataHash",
	}

	// add authentication in header
	ctx := context.Background()
	ctx = context.WithValue(ctx, auth.ContextKeyAuthenticatedAccountAddress{}, req.ProviderAddress)

	// request validation for provider data
	server := dataDealServiceServer{Service: suite.Svc, schema: validation.NewJSONSchema()}
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
		suite.OraclePubKey,
		crypto.KDFSHA256,
	)

	encryptedData, err := crypto.Encrypt(sharedKey, nil, jsonDataBz)
	suite.Require().NoError(err)

	req := &datadeal.ValidateDataRequest{
		DealId:          1,
		ProviderAddress: panacea.GetAddressFromPrivateKey(suite.providerAccPrivKey),
		EncryptedData:   encryptedData,
		DataHash:        "invalid data hash",
	}

	// add authentication in header
	ctx := context.Background()
	ctx = context.WithValue(ctx, auth.ContextKeyAuthenticatedAccountAddress{}, req.ProviderAddress)

	// request validation for provider data
	server := dataDealServiceServer{Service: suite.Svc}
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
		suite.OraclePrivKey.PubKey(),
		crypto.KDFSHA256,
	)

	encryptedData, err := crypto.Encrypt(sharedKey, nil, jsonDataBz)
	suite.Require().NoError(err)

	jsonData, err := jsoncanonicalizer.Transform(jsonDataBz)
	suite.Require().NoError(err)
	dataHash := sha256.Sum256(jsonData)

	req := &datadeal.ValidateDataRequest{
		DealId:          1,
		ProviderAddress: panacea.GetAddressFromPrivateKey(suite.providerAccPrivKey),
		EncryptedData:   encryptedData,
		DataHash:        hex.EncodeToString(dataHash[:]),
	}

	// add authentication in header
	ctx := context.Background()
	ctx = context.WithValue(ctx, auth.ContextKeyAuthenticatedAccountAddress{}, req.ProviderAddress)

	// request validation for provider data
	server := dataDealServiceServer{Service: suite.Svc, schema: validation.NewJSONSchema()}
	res, err := server.ValidateData(ctx, req)
	suite.Require().Nil(res)
	suite.Require().ErrorContains(err, "failed to validate data")
}

func (suite *dataDealServiceServerTestSuite) TestCanonicalJSON() {
	jsonDataBz := []byte(
		`
		{
			"invalid_key_name": "name",
			"invalid_key_description": "description",
			"invalid_key_body": [{ "type": "markdown", "attributes": { "value": "val1" } }]
		}
		`)

	// WhiteSpace before "name"
	jsonDataBzSpace := []byte(
		`
		{
			"invalid_key_name":  "name",
			"invalid_key_description": "description",
			"invalid_key_body": [{ "type": "markdown", "attributes": { "value": "val1" } }]
		}
		`)

	// WhiteSpace before "}" bracket
	jsonDataBzBracket := []byte(
		`
		{
			"invalid_key_name": "name",
			"invalid_key_description": "description",
			"invalid_key_body": [{ "type": "markdown", "attributes": { "value": "val1" } }]
		 }
		`)

	jsonDataBzOrder := []byte(
		`
		{
			"invalid_key_description": "description",
			"invalid_key_name": "name",
			"invalid_key_body": [{ "type": "markdown", "attributes": { "value": "val1" } }]
		}
		`)

	jsonData, err := jsoncanonicalizer.Transform(jsonDataBz)
	suite.Require().NoError(err)
	dataHash := sha256.Sum256(jsonData)

	jsonDataSpace, err := jsoncanonicalizer.Transform(jsonDataBzSpace)
	suite.Require().NoError(err)
	dataHashSpace := sha256.Sum256(jsonDataSpace)

	jsonDataOrder, err := jsoncanonicalizer.Transform(jsonDataBzOrder)
	suite.Require().NoError(err)
	dataHashOrder := sha256.Sum256(jsonDataOrder)

	jsonDataBracket, err := jsoncanonicalizer.Transform(jsonDataBzBracket)
	suite.Require().NoError(err)
	dataHashBracket := sha256.Sum256(jsonDataBracket)

	suite.Require().Equal(dataHash, dataHashOrder)
	suite.Require().Equal(dataHash, dataHashSpace)
	suite.Require().Equal(dataHashOrder, dataHashSpace)
	suite.Require().Equal(dataHash, dataHashBracket)
}
