package middleware

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"net/http"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/medibloc/panacea-oracle/panacea"
	log "github.com/sirupsen/logrus"
)

type jwtAuthMiddleware struct {
	panaceaQueryClient panacea.QueryClient

	// TODO: manage a nonce per account
}

func NewJWTAuthMiddleware(queryClient panacea.QueryClient) *jwtAuthMiddleware {
	return &jwtAuthMiddleware{
		panaceaQueryClient: queryClient,
	}
}

func (mw *jwtAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if len(authHeader) == 0 {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		jwtStr, err := parseBearerToken(authHeader)
		if err != nil {
			http.Error(w, "invalid bearer token", http.StatusUnauthorized)
			return
		}
		jwtBz := []byte(jwtStr)

		// parse JWT without signature verification to get payloads for retrieving an auth pubkey
		parsedJWT, err := jwt.ParseInsecure(jwtBz)
		if err != nil {
			http.Error(w, "invalid jwt", http.StatusUnauthorized)
			return
		}

		pubKey, err := mw.queryAccountPubKey(parsedJWT.Issuer())
		if err != nil {
			log.Error(err)
			http.Error(w, "cannot query account pubkey", http.StatusUnauthorized)
			return
		}

		_, err = jwt.Parse(jwtBz, jwt.WithKey(jwa.ES256K, pubKey))
		if err != nil {
			http.Error(w, "jwt signature verification failed", http.StatusUnauthorized)
			return
		}

		// pass the authenticated account address to next handlers
		newReq := r.WithContext(
			context.WithValue(r.Context(), ContextKeyAuthenticatedAccountAddress{}, parsedJWT.Issuer()),
		)

		next.ServeHTTP(w, newReq)
	})
}

func (mw *jwtAuthMiddleware) queryAccountPubKey(addr string) (*ecdsa.PublicKey, error) {
	height, err := mw.panaceaQueryClient.GetLastBlockHeight()
	if err != nil {
		return nil, fmt.Errorf("failed to get last block height. %v", err)
	}
	account, err := mw.panaceaQueryClient.GetAccount(height-1, addr)
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

func parseBearerToken(authHeader string) (string, error) {
	elems := strings.Split(authHeader, " ")
	if len(elems) != 2 || elems[0] != "Bearer" {
		return "", fmt.Errorf("invalid bearer token")
	}

	return elems[1], nil
}
