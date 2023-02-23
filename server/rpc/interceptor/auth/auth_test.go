package auth_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	didtypes "github.com/medibloc/panacea-core/v2/x/did/types"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	datadealtypes "github.com/medibloc/panacea-core/v2/x/datadeal/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/server/rpc/interceptor/auth"
	"github.com/stretchr/testify/require"

	tmtypes "github.com/tendermint/tendermint/types"
	"google.golang.org/grpc/metadata"
)

var (
	testPrivKey          = secp256k1.GenPrivKey()
	testOraclePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	testAccAddr          = "test-addr"
)

func TestAuthSuccess(t *testing.T) {
	jwt := testGenerateJWT(t, testAccAddr, testPrivKey, 10*time.Second)

	ctx := generateContextIncludeToken("bearer", string(jwt))

	ctx = testJWTInterceptor(
		t,
		ctx,
		&mockQueryClient{&mockAccount{}},
		"",
	)

	address, err := auth.GetRequestAddress(ctx)
	require.NoError(t, err)
	require.Equal(t, testAccAddr, address)
}

func TestMissingAuthorizationHeader(t *testing.T) {
	ctx := context.Background()
	testJWTInterceptor(
		t,
		ctx,
		&mockQueryClient{&mockAccount{}},
		"",
	)

	address, err := auth.GetRequestAddress(ctx)
	require.ErrorContains(t, err, "missing authorization header")
	require.Equal(t, "", address)
}

func TestInvalidBearerToken(t *testing.T) {
	jwt := testGenerateJWT(t, testAccAddr, testPrivKey, 10*time.Second)

	ctx := generateContextIncludeToken("bea123er", string(jwt))

	testJWTInterceptor(
		t,
		ctx,
		&mockQueryClient{&mockAccount{}},
		"",
	)

	address, err := auth.GetRequestAddress(ctx)
	require.ErrorContains(t, err, "missing authorization header")
	require.Equal(t, "", address)
}

func TestInvalidJWT(t *testing.T) {
	ctx := generateContextIncludeToken("bearer", "abcdef")

	testJWTInterceptor(
		t,
		ctx,
		&mockQueryClient{&mockAccount{}},
		"invalid bearer token",
	)
}

func TestAccountNotFound(t *testing.T) {
	jwt := testGenerateJWT(t, "dummy-account", testPrivKey, 10*time.Second)

	ctx := generateContextIncludeToken("bearer", string(jwt))

	testJWTInterceptor(
		t,
		ctx,
		&mockQueryClient{&mockAccount{}},
		"cannot query account pubkey",
	)
}

func TestAccountNoPubKey(t *testing.T) {
	jwt := testGenerateJWT(t, testAccAddr, testPrivKey, 10*time.Second)

	ctx := generateContextIncludeToken("bearer", string(jwt))

	testJWTInterceptor(
		t,
		ctx,
		&mockQueryClient{&mockAccountWithoutPubKey{}},
		"cannot query account pubkey",
	)
}

func TestSignatureVerificationFailure(t *testing.T) {
	jwt := testGenerateJWT(t, testAccAddr, secp256k1.GenPrivKey(), 10*time.Second)

	ctx := generateContextIncludeToken("bearer", string(jwt))

	testJWTInterceptor(
		t,
		ctx,
		&mockQueryClient{&mockAccount{}},
		"jwt signature verification failed",
	)
}

func testGenerateJWT(t *testing.T, issuer string, privKey *secp256k1.PrivKey, expiration time.Duration) []byte {
	now := time.Now().Truncate(time.Second)
	token, err := jwt.NewBuilder().
		Issuer(issuer).
		IssuedAt(now).
		NotBefore(now).
		Expiration(now.Add(expiration)).
		Build()
	require.NoError(t, err)

	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKey.Bytes())

	signedJWT, err := jwt.Sign(token, jwt.WithKey(jwa.ES256K, priv.ToECDSA()))
	require.NoError(t, err)

	return signedJWT
}

func generateContextIncludeToken(tokenType, jwt string) context.Context {
	ctx := context.Background()

	m := map[string]string{
		"authorization": fmt.Sprintf("%s %s", tokenType, jwt),
	}

	md := metadata.New(m)
	ctx = metadata.NewIncomingContext(ctx, md)
	return ctx
}

func testJWTInterceptor(
	t *testing.T,
	ctx context.Context,
	queryClient panacea.QueryClient,
	errMsg string,
) context.Context {
	interceptor := auth.NewJWTAuthInterceptor(queryClient)

	ctx, err := interceptor.Interceptor(ctx)
	if errMsg != "" {
		require.ErrorContains(t, err, errMsg)
	} else {
		require.NoError(t, err)
	}

	return ctx
}

type mockQueryClient struct {
	account authtypes.AccountI
}

func (c *mockQueryClient) GetConsent(_ context.Context, _ uint64, _ string) (*datadealtypes.Consent, error) {
	return nil, nil
}

func (c *mockQueryClient) GetOracleRegistration(_ context.Context, uniqueID, oracleAddr string) (*oracletypes.OracleRegistration, error) {
	return nil, nil
}

func (c *mockQueryClient) GetLightBlock(height int64) (*tmtypes.LightBlock, error) {
	return nil, nil
}

func (c *mockQueryClient) GetCdc() *codec.ProtoCodec {
	return nil
}

func (c *mockQueryClient) GetChainID() string {
	return ""
}

func (c *mockQueryClient) Close() error {
	return nil
}

func (c *mockQueryClient) GetAccount(_ context.Context, address string) (authtypes.AccountI, error) {
	if address != testAccAddr {
		return nil, fmt.Errorf("address not found: %v", address)
	}
	return c.account, nil
}

func (c *mockQueryClient) GetDID(_ context.Context, _ string) (*didtypes.DIDDocumentWithSeq, error) {
	return nil, nil
}
func (c *mockQueryClient) GetOracleUpgrade(_ context.Context, _, _ string) (*oracletypes.OracleUpgrade, error) {
	return nil, nil
}

func (c *mockQueryClient) GetDeal(_ context.Context, _ uint64) (*datadealtypes.Deal, error) {
	return nil, nil
}

func (c *mockQueryClient) GetOracleUpgradeInfo(_ context.Context) (*oracletypes.OracleUpgradeInfo, error) {
	return nil, nil
}

func (c *mockQueryClient) GetOracle(_ context.Context, _ string) (*oracletypes.Oracle, error) {
	return nil, nil
}

func (c *mockQueryClient) VerifyTrustedBlockInfo(_ int64, _ []byte) error {
	return nil
}

type mockAccount struct{}

func (a *mockAccount) Reset() {
}

func (a *mockAccount) String() string {
	return ""
}

func (a *mockAccount) ProtoMessage() {
}

func (a *mockAccount) GetAddress() sdk.AccAddress {
	return nil
}

func (a *mockAccount) SetAddress(address sdk.AccAddress) error {
	return nil
}

func (a *mockAccount) GetPubKey() cryptotypes.PubKey {
	return testPrivKey.PubKey()
}

func (a *mockAccount) SetPubKey(key cryptotypes.PubKey) error {
	return nil
}

func (a *mockAccount) GetAccountNumber() uint64 {
	return 0
}

func (a *mockAccount) SetAccountNumber(u uint64) error {
	return nil
}

func (a *mockAccount) GetSequence() uint64 {
	return 0
}

func (a *mockAccount) SetSequence(u uint64) error {
	return nil
}

func (c *mockQueryClient) GetOracleParamsPublicKey(_ context.Context) (*btcec.PublicKey, error) {
	return testOraclePrivKey.PubKey(), nil
}

type mockAccountWithoutPubKey struct {
	mockAccount
}

func (a *mockAccountWithoutPubKey) GetPubKey() cryptotypes.PubKey {
	return nil
}

func (c *mockQueryClient) GetLastBlockHeight(_ context.Context) (int64, error) {
	return 0, nil
}
