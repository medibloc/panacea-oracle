package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/medibloc/panacea-oracle/server/service"
	"github.com/medibloc/panacea-oracle/types"
	"github.com/medibloc/panacea-oracle/validation"
	log "github.com/sirupsen/logrus"
	"net/http"
	"regexp"
	"strings"
)

const (
	prefixType = "Signature"

	EsSha256 = "es256k1-sha256"
)

var authorizationHeaders = []string{types.AuthAlgorithmHeaderKey, types.AuthKeyIDHeaderKey, types.AuthNonceHeaderKey, types.AuthSignatureHeaderKey}
var authenticateHeaders = []string{types.AuthAlgorithmHeaderKey, types.AuthKeyIDHeaderKey, types.AuthNonceHeaderKey}
var wantedAlgorithms = []string{EsSha256}

type AuthenticationMiddleware struct {
	service *service.Service
	url     map[string][]string
}

func NewMiddleware(svc *service.Service) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{
		service: svc,
		url:     make(map[string][]string),
	}
}

func (amw *AuthenticationMiddleware) AddURL(path string, methods ...string) {
	// Router is only used to convert path to regex.
	pathRegex, err := mux.NewRouter().Path(path).GetPathRegexp()
	if err != nil {
		panic(err)
	}
	m, ok := amw.url[pathRegex]
	if !ok {
		amw.url[pathRegex] = methods
	} else {
		amw.url[pathRegex] = append(m, methods...)
	}
}

// Middleware supports signature authentication and responds to the client by generating information necessary for authentication.
func (amw *AuthenticationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if it is not an authentication URL, the authentication check is not performed.
		if !amw.isAuthenticationURL(r) {
			next.ServeHTTP(w, r)
			return
		}

		sigAuthParts, err := ParseSignatureAuthorizationParts(r.Header.Get("Authorization"))
		//signatureAuthentication, err := parseSignatureAuthentication(r)
		if err != nil {
			// result code to 400. Bad request
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = basicValidate(sigAuthParts)
		if err != nil {
			// result code to 400. Bad request
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if sigAuthParts[types.AuthNonceHeaderKey] == "" {
			err = amw.generateAuthenticationAndSetHeader(w, sigAuthParts)
			if err != nil {
				log.Error("failed to generate authentication", err)
				http.Error(w, "failed to generate authentication", http.StatusInternalServerError)
				return
			}
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		cache := amw.service.Cache
		auth := cache.Get(sigAuthParts[types.AuthKeyIDHeaderKey], sigAuthParts[types.AuthNonceHeaderKey])
		if auth == nil {
			// expired
			err = amw.generateAuthenticationAndSetHeader(w, sigAuthParts)
			if err != nil {
				log.Error("failed to generate authentication", err)
				http.Error(w, "failed to generate authentication", http.StatusInternalServerError)
				return
			}
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		nonce := sigAuthParts[types.AuthNonceHeaderKey]
		signature, err := base64.StdEncoding.DecodeString(sigAuthParts[types.AuthSignatureHeaderKey])
		if err != nil {
			http.Error(w, "failed to decode signature", http.StatusBadRequest)
			return
		}

		requesterAddress := sigAuthParts[types.AuthKeyIDHeaderKey]
		pubKey, err := amw.service.PanaceaClient.GetPubKey(requesterAddress)
		if err != nil {
			log.Error(fmt.Sprintf("failed to get the account's public key account: %s", requesterAddress), err)
			http.Error(w, "failed to get the account's public key", http.StatusInternalServerError)
			return
		}

		if !pubKey.VerifySignature([]byte(nonce), signature) {
			http.Error(w, "failed to verification signature", http.StatusBadRequest)
			return
		}

		cache.Remove(sigAuthParts[types.AuthKeyIDHeaderKey], nonce)

		context.Set(r, types.RequesterAddressKey, requesterAddress)
		next.ServeHTTP(w, r)
	})
}

func (amw *AuthenticationMiddleware) isAuthenticationURL(r *http.Request) bool {
	for path, methods := range amw.url {
		ok, _ := regexp.MatchString(path, r.URL.Path)
		if ok && validation.Contains(methods, r.Method) {
			return true
		}
	}
	return false
}

// ParseSignatureAuthorizationParts parses Authorization value in Header according to `Signature` type.
func ParseSignatureAuthorizationParts(auth string) (map[string]string, error) {
	if len(auth) < len(prefixType) || !strings.EqualFold(auth[:len(prefixType)], prefixType) {
		return nil, errors.New("not supported auth type")
	}

	headers := strings.Split(auth[len(prefixType):], ",")
	parts := make(map[string]string, len(authorizationHeaders))
	for _, header := range headers {
		for _, w := range authorizationHeaders {
			if strings.Contains(header, w) {
				parts[w] = strings.Split(header, `"`)[1]
			}
		}
	}

	return parts, nil
}

func basicValidate(sigAuthParts map[string]string) error {
	if !validation.Contains(wantedAlgorithms, sigAuthParts[types.AuthAlgorithmHeaderKey]) {
		return errors.New(fmt.Sprintf("is not supported value. (Algorithm: %s)", sigAuthParts[types.AuthAlgorithmHeaderKey]))
	} else if sigAuthParts[types.AuthKeyIDHeaderKey] == "" {
		return errors.New("'KeyId' cannot be empty")
	}
	return nil
}

// generateAuthenticationAndSetHeader creates authentication request information and puts it in the header of the response.
func (amw *AuthenticationMiddleware) generateAuthenticationAndSetHeader(w http.ResponseWriter, sigAuthParts map[string]string) error {
	err := amw.generateAuthentication(sigAuthParts)
	if err != nil {
		return err
	}
	auth := makeAuthenticationHeader(sigAuthParts)

	w.Header().Set("WWW-Authenticate", auth)

	return nil
}

// generateAuthentication creates a nonce and stores it in the cache.
func (amw *AuthenticationMiddleware) generateAuthentication(sigAuthParts map[string]string) error {
	err := setNewNonce(sigAuthParts)
	if err != nil {
		return err
	}

	cache := amw.service.Cache
	err = cache.Set(sigAuthParts[types.AuthKeyIDHeaderKey], sigAuthParts[types.AuthNonceHeaderKey], sigAuthParts)
	if err != nil {
		return err
	}

	return nil
}

func setNewNonce(sigAuthParts map[string]string) error {
	randomBytes := make([]byte, 24)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return err
	}

	sigAuthParts[types.AuthNonceHeaderKey] = base64.StdEncoding.EncodeToString(randomBytes)

	return nil
}

func makeAuthenticationHeader(sigAuthParts map[string]string) string {
	var authenticate []string
	for _, h := range authenticateHeaders {
		authenticate = append(authenticate, fmt.Sprintf("%s=\"%s\"", h, sigAuthParts[h]))
	}

	return prefixType + " " + strings.Join(authenticate, ", ")
}
