package middleware_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-oracle/server/middleware"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"
)

var (
	testPrivKey = secp256k1.GenPrivKey()
	testAccAddr = "test-addr"

	testHandler = middleware.NewJWTAuthMiddleware(&mockQueryClient{}).Middleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For tests, return OK only if the request is requested by the 'testAccAddr'
			accAddr := r.Context().Value(middleware.ContextKeyAuthenticatedAccountAddress{}).(string)
			if accAddr == testAccAddr {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusNotAcceptable)
			}
		}),
	)
)

func TestAuthSuccess(t *testing.T) {
	jwt := testGenerateJWT(t, testAccAddr, testPrivKey, 10*time.Second)
	testHTTPRequest(
		t,
		fmt.Sprintf("Bearer %s", string(jwt)),
		http.StatusOK,
		"",
	)
}

func TestMissingAuthorizationHeader(t *testing.T) {
	testHTTPRequest(
		t,
		"",
		http.StatusUnauthorized,
		"missing authorization header",
	)
}

func TestInvalidBearerToken(t *testing.T) {
	jwt := testGenerateJWT(t, testAccAddr, testPrivKey, 10*time.Second)
	testHTTPRequest(
		t,
		fmt.Sprintf("Bea123er %s", string(jwt)),
		http.StatusUnauthorized,
		"invalid bearer token",
	)
}

func TestInvalidJWT(t *testing.T) {
	testHTTPRequest(
		t,
		"Bearer abcdef",
		http.StatusUnauthorized,
		"invalid jwt",
	)
}

func TestAccountNotFound(t *testing.T) {
	jwt := testGenerateJWT(t, "dummy-account", testPrivKey, 10*time.Second)
	testHTTPRequest(
		t,
		fmt.Sprintf("Bearer %s", string(jwt)),
		http.StatusUnauthorized,
		"cannot query account pubkey",
	)
}

func TestSignatureVerificationFailure(t *testing.T) {
	jwt := testGenerateJWT(t, testAccAddr, secp256k1.GenPrivKey(), 10*time.Second)
	testHTTPRequest(
		t,
		fmt.Sprintf("Bearer %s", string(jwt)),
		http.StatusUnauthorized,
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

func testHTTPRequest(t *testing.T, authorizationHeader string, statusCode int, errMsg string) {
	req := httptest.NewRequest("GET", "http://test.com", nil)
	req.Header.Set("Authorization", authorizationHeader)

	w := httptest.NewRecorder()
	testHandler.ServeHTTP(w, req)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, statusCode, resp.StatusCode)
	if errMsg != "" {
		require.Equal(t, errMsg+"\n", string(body))
	}
}

//// Mocks //////////////////////////////////////////////////////////////

type mockQueryClient struct{}

func (c *mockQueryClient) GetOracleRegistration(uniqueID, oracleAddr string) (*oracletypes.OracleRegistration, error) {
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

func (c *mockQueryClient) GetAccount(address string) (authtypes.AccountI, error) {
	if address != testAccAddr {
		return nil, fmt.Errorf("address not found: %v", address)
	}
	return &mockAccount{}, nil
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
