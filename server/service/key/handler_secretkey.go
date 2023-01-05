package key

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/btcsuite/btcd/btcec"
	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/medibloc/panacea-oracle/server/middleware"
	log "github.com/sirupsen/logrus"
)

func (svc *combinedKeyService) GetSecretKey(w http.ResponseWriter, r *http.Request) {
	queryClient := svc.QueryClient()
	oraclePrivKey := svc.OraclePrivKey()

	queryParams := r.URL.Query()

	dealIDStr := queryParams.Get("deal-id")
	dealID, err := strconv.ParseUint(dealIDStr, 10, 64)
	if err != nil {
		log.Errorf("failed to parse deal ID: %s", err.Error())
		http.Error(w, "failed to parse deal ID", http.StatusBadRequest)
		return
	}

	dataHashStr := queryParams.Get("data-hash")
	dataHash, err := hex.DecodeString(dataHashStr)
	if err != nil {
		log.Errorf("failed to decode dataHash: %s", err.Error())
		http.Error(w, "failed to decode dataHash", http.StatusBadRequest)
		return
	}

	// Check the address of the requested consumer
	accAddr := r.Context().Value(middleware.ContextKeyAuthenticatedAccountAddress{}).(string)
	deal, err := queryClient.GetDeal(r.Context(), dealID)
	if err != nil {
		log.Errorf("failed to get deal(%d): %s", dealID, err.Error())
		http.Error(w, "failed to get deal", http.StatusNotFound)
		return
	}

	if accAddr != deal.ConsumerAddress {
		log.Error("only consumer request secret key")
		http.Error(w, "only consumer request secret key", http.StatusForbidden)
		return
	}

	// Check if the consent has been submitted
	_, err = queryClient.GetConsent(r.Context(), dealID, dataHashStr)
	if err != nil {
		log.Errorf("failed to get consent(dealID: %d, dataHash %s): %s", dealID, dataHashStr, err.Error())
		http.Error(w, "failed to get consent", http.StatusNotFound)
		return
	}

	// make encrypted secret key using consumer public key
	consumerAcc, err := queryClient.GetAccount(r.Context(), deal.ConsumerAddress)
	if err != nil {
		log.Errorf("failed to get consumer account: %s", err.Error())
		http.Error(w, "failed to get consumer account", http.StatusNotFound)
		return
	}
	consumerPubKeyBz := consumerAcc.GetPubKey().Bytes()
	consumerPubKey, err := btcec.ParsePubKey(consumerPubKeyBz, btcec.S256())
	if err != nil {
		log.Errorf("failed to parse consumer public key: %s", err.Error())
		http.Error(w, "failed to parse consumer public key", http.StatusInternalServerError)
		return
	}

	sharedKey := crypto.DeriveSharedKey(oraclePrivKey, consumerPubKey, crypto.KDFSHA256)

	secretKey := GetCombinedKey(oraclePrivKey.Serialize(), dealID, dataHash)
	encryptedSecretKey, err := crypto.Encrypt(sharedKey, nil, secretKey)
	if err != nil {
		log.Errorf("failed to encrypt secret key with shared key: %s", err.Error())
		http.Error(w, "failed to encrypt secret key with shared key", http.StatusInternalServerError)
		return
	}

	// make response
	var response secretKeyResponse
	response.EncryptedSecretKey = encryptedSecretKey
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Errorf("failed to marshal response: %s", err.Error())
		http.Error(w, "failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)
	if err != nil {
		log.Errorf("failed to write response: %s", err.Error())
		return
	}
}
