package auth_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/server/rpc/interceptor/auth"
	"github.com/medibloc/panacea-oracle/types/test_utils/mocks"
	"github.com/stretchr/testify/require"

	"google.golang.org/grpc/metadata"
)

var (
	testPrivKey = secp256k1.GenPrivKey()
	testAccAddr = "test-addr"
)

func TestAuthSuccess(t *testing.T) {
	jwt := testGenerateJWT(t, testAccAddr, testPrivKey, 10*time.Second)

	ctx := generateContextIncludeToken("bearer", string(jwt))

	ctx = testJWTInterceptor(
		t,
		ctx,
		&mocks.MockQueryClient{
			Account: mocks.NewMockAccount(testPrivKey.PubKey()),
		},
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
		&mocks.MockQueryClient{
			Account: mocks.NewMockAccount(testPrivKey.PubKey()),
		},
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
		&mocks.MockQueryClient{
			Account: mocks.NewMockAccount(testPrivKey.PubKey()),
		},
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
		&mocks.MockQueryClient{
			Account: mocks.NewMockAccount(testPrivKey.PubKey()),
		},
		"invalid bearer token",
	)
}

func TestAccountNotFound(t *testing.T) {
	jwt := testGenerateJWT(t, "dummy-account", testPrivKey, 10*time.Second)

	ctx := generateContextIncludeToken("bearer", string(jwt))

	testJWTInterceptor(
		t,
		ctx,
		&mocks.MockQueryClient{
			AccountError: fmt.Errorf("not found account"),
		},
		"cannot query account pubkey",
	)
}

func TestAccountNoPubKey(t *testing.T) {
	jwt := testGenerateJWT(t, testAccAddr, testPrivKey, 10*time.Second)

	ctx := generateContextIncludeToken("bearer", string(jwt))

	account := mocks.NewMockAccount(testPrivKey.PubKey())
	account.PubKey = nil
	testJWTInterceptor(
		t,
		ctx,
		&mocks.MockQueryClient{
			Account: account,
		},
		"cannot query account pubkey",
	)
}

func TestSignatureVerificationFailure(t *testing.T) {
	jwt := testGenerateJWT(t, testAccAddr, secp256k1.GenPrivKey(), 10*time.Second)

	ctx := generateContextIncludeToken("bearer", string(jwt))

	testJWTInterceptor(
		t,
		ctx,
		&mocks.MockQueryClient{
			Account: mocks.NewMockAccount(testPrivKey.PubKey()),
		},
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
