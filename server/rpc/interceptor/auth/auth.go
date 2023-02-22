package auth

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/medibloc/panacea-oracle/panacea"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type jwtAuthInterceptor struct {
	panaceaQueryClient panacea.QueryClient
}

func NewJWTAuthInterceptor(queryClient panacea.QueryClient) *jwtAuthInterceptor {
	return &jwtAuthInterceptor{queryClient}
}

func (ic *jwtAuthInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return grpc_auth.UnaryServerInterceptor(ic.Interceptor)
}

func (ic *jwtAuthInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return grpc_auth.StreamServerInterceptor(ic.Interceptor)
}

// Interceptor defines that if there is no authentication value, it returns without taking any action.
// A service that uses this should respond with an error when there is no authenticated user.
// The token is normally included in the header, but responds with an error if validation fails.
func (ic *jwtAuthInterceptor) Interceptor(ctx context.Context) (context.Context, error) {
	log.Debug("Call jwt interceptor")

	jwtTokenStr, err := grpc_auth.AuthFromMD(ctx, "bearer")

	if err != nil {
		log.Debugf("failed to jwt token from header. %v", err.Error())
		return ctx, nil
	}

	log.Debugf("jwt token: %s", jwtTokenStr)

	jwtBz := []byte(jwtTokenStr)

	parsedJWT, err := jwt.ParseInsecure(jwtBz)
	if err != nil {
		return nil, fmt.Errorf("invalid bearer token. %w", err)
	}

	pubKey, err := ic.queryAccountPubKey(ctx, parsedJWT.Issuer())
	if err != nil {
		return nil, fmt.Errorf("cannot query account pubkey. %w", err)
	}

	_, err = jwt.Parse(jwtBz, jwt.WithKey(jwa.ES256K, pubKey))
	if err != nil {
		return nil, fmt.Errorf("jwt signature verification failed. %w", err)
	}

	newCtx := context.WithValue(ctx, ContextKeyAuthenticatedAccountAddress{}, parsedJWT.Issuer())

	return newCtx, nil
}

func (ic *jwtAuthInterceptor) queryAccountPubKey(ctx context.Context, addr string) (*ecdsa.PublicKey, error) {
	account, err := ic.panaceaQueryClient.GetAccount(ctx, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to query account: %w", err)
	}

	pubKey := account.GetPubKey()
	if pubKey == nil {
		return nil, fmt.Errorf("no pubkey registered to the account yet")
	}

	parsedPubKey, err := btcec.ParsePubKey(pubKey.Bytes(), btcec.S256())
	if err != nil {
		return nil, fmt.Errorf("failed to parse account pubkey: %w", err)
	}

	return parsedPubKey.ToECDSA(), nil
}

func GetRequestAddress(ctx context.Context) (string, error) {
	reqAddr := ctx.Value(ContextKeyAuthenticatedAccountAddress{})
	if reqAddr == nil {
		return "", fmt.Errorf("missing authorization header")
	}
	return reqAddr.(string), nil
}
