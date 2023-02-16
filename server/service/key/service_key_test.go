package key

import (
	"context"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	datadealtypes "github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/medibloc/panacea-oracle/panacea"
	key "github.com/medibloc/panacea-oracle/pb/key/v0"
	"github.com/medibloc/panacea-oracle/server/rpc/interceptor/auth"
	"github.com/medibloc/panacea-oracle/types/test_utils/mocks"
	"github.com/stretchr/testify/suite"
)

type secretKeyServiceTestSuite struct {
	suite.Suite

	svc *mocks.MockService

	oraclePrivKey *btcec.PrivateKey

	consumerAccPrivKey secp256k1.PrivKey
	consumerAccPubKey  cryptotypes.PubKey
	consumerAcc        authtypes.AccountI
	consumerAddress    string
}

func TestCombinedKeyServiceTestSuite(t *testing.T) {
	suite.Run(t, &secretKeyServiceTestSuite{})
}

func (suite *secretKeyServiceTestSuite) BeforeTest(_, _ string) {
	suite.oraclePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	suite.consumerAccPrivKey = *secp256k1.GenPrivKey()
	suite.consumerAccPubKey = suite.consumerAccPrivKey.PubKey()
	suite.consumerAcc = mocks.NewMockAccount(suite.consumerAccPubKey)
	suite.consumerAddress = panacea.GetAddressFromPrivateKey(suite.consumerAccPrivKey)
	suite.svc = &mocks.MockService{
		OraclePrivKey: suite.oraclePrivKey,
		QueryClient: &mocks.MockQueryClient{
			Deal:    &datadealtypes.Deal{},
			Consent: &datadealtypes.Consent{},
			Account: suite.consumerAcc,
		},
	}

}

func (suite *secretKeyServiceTestSuite) TestGetSecretKey() {
	combinedKeyService := secretKeyService{Service: suite.svc}
	data := "my_data"
	dataHash := crypto.KDFSHA256([]byte(data))

	suite.svc.QueryClient.Deal.ConsumerAddress = suite.consumerAddress

	req := &key.GetSecretKeyRequest{
		DealId:   1,
		DataHash: dataHash,
	}

	ctx := context.Background()
	ctx = context.WithValue(
		ctx,
		auth.ContextKeyAuthenticatedAccountAddress{},
		suite.consumerAddress,
	)

	res, err := combinedKeyService.GetSecretKey(ctx, req)
	suite.Require().NoError(err)

	consumerPrivKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), suite.consumerAccPrivKey.Bytes())
	sharedKey := crypto.DeriveSharedKey(
		consumerPrivKey,
		suite.oraclePrivKey.PubKey(),
		crypto.KDFSHA256,
	)

	secretKey, err := crypto.Decrypt(sharedKey, nil, res.EncryptedSecretKey)
	suite.Require().NoError(err)

	suite.Require().Equal(
		GetSecretKey(suite.oraclePrivKey.Serialize(), req.DealId, req.DataHash),
		secretKey,
	)
}

func (suite *secretKeyServiceTestSuite) TestGetSecretKeyNotExistAuthentication() {
	combinedKeyService := secretKeyService{Service: suite.svc}
	data := "my_data"
	dataHash := crypto.KDFSHA256([]byte(data))

	suite.svc.QueryClient.Deal.ConsumerAddress = suite.consumerAddress

	req := &key.GetSecretKeyRequest{
		DealId:   1,
		DataHash: dataHash,
	}

	ctx := context.Background()

	res, err := combinedKeyService.GetSecretKey(ctx, req)
	suite.Require().Nil(res)
	suite.Require().ErrorContains(err, "failed to get request address")
}

func (suite *secretKeyServiceTestSuite) TestGetSecretKeyNotSameRequesterAndDealsConsumer() {
	combinedKeyService := secretKeyService{Service: suite.svc}
	data := "my_data"
	dataHash := crypto.KDFSHA256([]byte(data))

	suite.svc.QueryClient.Deal.ConsumerAddress = suite.consumerAddress

	req := &key.GetSecretKeyRequest{
		DealId:   1,
		DataHash: dataHash,
	}

	ctx := context.Background()
	ctx = context.WithValue(
		ctx,
		auth.ContextKeyAuthenticatedAccountAddress{},
		panacea.GetAddressFromPrivateKey(*secp256k1.GenPrivKey()),
	)

	res, err := combinedKeyService.GetSecretKey(ctx, req)
	suite.Require().Nil(res)
	suite.Require().ErrorContains(err, "only consumer request secret key")
}
