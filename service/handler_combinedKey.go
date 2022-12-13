package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/mux"
	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/medibloc/panacea-oracle/server/middleware"
	log "github.com/sirupsen/logrus"
)

type Response struct {
	EncryptedCombinedKey []byte `json:"encrypted-combined-key"`
}

func (svc *Service) GetCombinedKey(w http.ResponseWriter, r *http.Request) {
	queryClient := svc.QueryClient()
	oraclePrivKey := svc.OraclePrivKey()

	dealIDStr := mux.Vars(r)["dealId"]
	dealID, err := strconv.ParseUint(dealIDStr, 10, 64)
	if err != nil {
		log.Errorf("failed to parse deal ID: %s", err.Error())
		http.Error(w, "failed to parse deal ID", http.StatusBadRequest)
		return
	}

	dataHashStr := mux.Vars(r)["dataHash"]
	dataHash, err := hex.DecodeString(dataHashStr)
	if err != nil {
		log.Errorf("failed to decode dataHash: %s", err.Error())
		http.Error(w, "failed to decode dataHash", http.StatusBadRequest)
		return
	}

	// Check the address of the requested consumer
	accAddr := r.Context().Value(middleware.ContextKeyAuthenticatedAccountAddress{}).(string)
	deal, err := queryClient.GetDeal(dealID)
	if err != nil {
		log.Errorf("failed to get deal(%d): %s", dealID, err.Error())
		http.Error(w, "failed to get deal", http.StatusBadRequest)
		return
	}

	if accAddr != deal.ConsumerAddress {
		log.Error("only consumer request combined key")
		http.Error(w, "only consumer request combined key", http.StatusBadRequest)
		return
	}

	// Check if the certificate of data has been submitted
	_, err = queryClient.GetCertificate(dealID, dataHashStr)
	if err != nil {
		log.Errorf("failed to get certificate(dealID: %d, dataHash %s): %s", dealID, dataHashStr, err.Error())
		http.Error(w, "failed to get certificate", http.StatusBadRequest)
		return
	}

	// make encrypted combined key using consumer public key
	consumerAcc, err := queryClient.GetAccount(deal.ConsumerAddress)
	if err != nil {
		log.Errorf("failed to get deal(%d): %s", dealID, err.Error())
		http.Error(w, "failed to get deal", http.StatusBadRequest)
		return
	}
	consumerPubKeyBz := consumerAcc.GetPubKey().Bytes()
	consumerPubKey, err := btcec.ParsePubKey(consumerPubKeyBz, btcec.S256())
	if err != nil {
		log.Errorf("failed to parse consumer public key: %s", err.Error())
		http.Error(w, "failed to parse consumer public key", http.StatusBadRequest)
		return
	}

	sharedKey := crypto.DeriveSharedKey(oraclePrivKey, consumerPubKey, crypto.KDFSHA256)

	combinedKey := getCombinedKey(oraclePrivKey.Serialize(), dealID, dataHash)
	encryptedCombinedKey, err := crypto.EncryptWithAES256(sharedKey, combinedKey[:])
	if err != nil {
		log.Errorf("failed to encrypt combined key with shared key: %s", err.Error())
		http.Error(w, "failed to encrypt combined key with shared key", http.StatusInternalServerError)
		return
	}

	// make response
	var response Response
	response.EncryptedCombinedKey = encryptedCombinedKey
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Errorf("failed to marshal response: %s", err.Error())
		http.Error(w, "failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(jsonResponse)
	if err != nil {
		log.Errorf("failed to write response: %s", err.Error())
		return
	}
}

func getCombinedKey(oraclePrivKey []byte, dealID uint64, dataHash []byte) [sha256.Size]byte {
	tmp := append(oraclePrivKey, sdk.Uint64ToBigEndian(dealID)...)
	return sha256.Sum256(append(tmp, dataHash...))
}
