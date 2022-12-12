package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/mux"
	"github.com/medibloc/panacea-oracle/server/middleware"
	log "github.com/sirupsen/logrus"
)

type Response struct {
	CombinedKey []byte `json:"combinedKey"`
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
	var dataHash32 [sha256.Size]byte
	copy(dataHash32[:], dataHash)

	if err != nil {
		log.Errorf("failed to decode dataHash: %s", err.Error())
		http.Error(w, "failed to decode dataHash", http.StatusBadRequest)
		return
	}

	// Check if the certificate of data has been submitted
	certificate, err := queryClient.GetCertificate(dealID, dataHashStr)
	if err != nil {
		log.Errorf("failed to get certificate: %s", err.Error())
		http.Error(w, "failed to get certificate", http.StatusBadRequest)
		return
	}

	// Check the address of the requested consumer
	accAddr := r.Context().Value(middleware.ContextKeyAuthenticatedAccountAddress{}).(string)

	if accAddr != certificate.UnsignedCertificate.ProviderAddress {
		log.Error("only consumer request combined key")
		http.Error(w, "only consumer request combined key", http.StatusBadRequest)
		return
	}

	// response combinedKey
	combinedKey := getCombinedKey(oraclePrivKey.Serialize(), dealID, dataHash32)
	var response Response
	response.CombinedKey = combinedKey[:]
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Errorf("failed to marshal payload: %s", err.Error())
		http.Error(w, "failed to marshal payload", http.StatusInternalServerError)
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

func getCombinedKey(oraclePrivKey []byte, dealID uint64, dataHash [sha256.Size]byte) [sha256.Size]byte {
	tmp := append(oraclePrivKey, sdk.Uint64ToBigEndian(dealID)...)
	return sha256.Sum256(append(tmp, dataHash[:]...))
}
