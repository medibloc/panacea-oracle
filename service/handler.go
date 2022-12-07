package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/mux"
	datadealtypes "github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/medibloc/panacea-oracle/validation"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type ValidateDataReq struct {
	ProviderAddress string `json:"provider_address"`
	EncryptedData   string `json:"encrypted_data"`
	DataHash        string `json:"data_hash"`
}

func (svc *Service) ValidateData(w http.ResponseWriter, r *http.Request) {
	queryClient := svc.QueryClient()
	oraclePrivKey := svc.OraclePrivKey()

	// Read a data from request body
	dealIDStr := mux.Vars(r)["dealId"]
	dealID, err := strconv.ParseUint(dealIDStr, 10, 64)
	if err != nil {
		log.Errorf("failed to parse deal ID: %s", err.Error())
		http.Error(w, "failed to parse deal ID", http.StatusBadRequest)
		return
	}

	var reqBody ValidateDataReq

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Errorf("failed to decode request body: %s", err.Error())
		http.Error(w, "failed to decode request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Decrypt data
	encryptedDataBz, err := hex.DecodeString(reqBody.EncryptedData)
	if err != nil {
		log.Errorf("failed to decode encrypted data: %s", err.Error())
		http.Error(w, "failed to decode encrypted data", http.StatusBadRequest)
		return
	}

	providerAcc, err := queryClient.GetAccount(reqBody.ProviderAddress)
	if err != nil {
		log.Errorf("failed to get provider's account: %s", err.Error())
		http.Error(w, "failed to get provider's account", http.StatusBadRequest)
		return
	}
	providerPubKeyBytes := providerAcc.GetPubKey().Bytes()

	providerPubKey, err := btcec.ParsePubKey(providerPubKeyBytes, btcec.S256())
	if err != nil {
		log.Errorf("failed to parse provider's public key: %s", err.Error())
		http.Error(w, "failed to parse provider's public key", http.StatusBadRequest)
		return
	}

	decryptSharedKey := crypto.DeriveSharedKey(oraclePrivKey, providerPubKey, crypto.KDFSHA256)

	decryptedData, err := crypto.Decrypt(decryptSharedKey, nil, encryptedDataBz)
	if err != nil {
		log.Errorf("failed to decrypt data: %s", err.Error())
		http.Error(w, "failed to decrypt data", http.StatusBadRequest)
		return
	}

	// Validate data
	dataHash := sha256.Sum256(decryptedData)
	dataHashStr := hex.EncodeToString(dataHash[:])
	if reqBody.DataHash != dataHashStr {
		log.Errorf("data hash mismatch")
		http.Error(w, "data hash mismatch", http.StatusBadRequest)
		return
	}

	deal, err := queryClient.GetDeal(dealID)
	if err != nil {
		log.Errorf("failed to get deal(%d): %s", dealID, err.Error())
		http.Error(w, "failed to get deal", http.StatusBadRequest)
		return
	}

	if err := validation.ValidateJSONSchemata(decryptedData, deal.DataSchema); err != nil {
		log.Errorf("failed to validate data: %s", err.Error())
		http.Error(w, "failed to validate data", http.StatusBadRequest)
		return
	}

	// Re-encrypt data using a combined key
	combinedKey := getCombinedKey(oraclePrivKey.Serialize(), dealID, dataHash)
	reEncryptedData, err := crypto.Encrypt(combinedKey[:], nil, decryptedData)
	if err != nil {
		log.Errorf("failed to re-encrypt data with the combined key: %s", err.Error())
		http.Error(w, "failed to re-encrypt data with the combined key", http.StatusInternalServerError)
		return
	}

	// Put data into IPFS
	cid, err := svc.IPFS().Add(reEncryptedData)
	if err != nil {
		log.Errorf("failed to store data to IPFS: %s", err.Error())
		http.Error(w, "failed to store data to IPFS", http.StatusInternalServerError)
	}

	// Issue a certificate to the client
	unsignedDataCert := &datadealtypes.UnsignedCertificate{
		Cid:             cid,
		OracleAddress:   svc.OracleAcc().GetAddress(),
		DealId:          dealID,
		ProviderAddress: reqBody.ProviderAddress,
		DataHash:        dataHashStr,
	}

	key, _ := btcec.PrivKeyFromBytes(btcec.S256(), oraclePrivKey.Serialize())

	marshaledDataCert, err := json.Marshal(unsignedDataCert)
	if err != nil {
		log.Errorf("failed to marshal data certificate: %s", err.Error())
		http.Error(w, "failed to marshal data certificate", http.StatusInternalServerError)
		return
	}

	sig, err := key.Sign(marshaledDataCert)
	if err != nil {
		log.Errorf("failed to create signature of data certificate: %s", err.Error())
		http.Error(w, "failed to create signature of data certificate", http.StatusInternalServerError)
		return
	}

	payload := datadealtypes.Certificate{
		UnsignedCertificate: unsignedDataCert,
		Signature:           sig.Serialize(),
	}

	marshaledPayload, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("failed to marshal payload: %s", err.Error())
		http.Error(w, "failed to marshal payload", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(marshaledPayload)
}

func getCombinedKey(oraclePrivKey []byte, dealID uint64, dataHash [sha256.Size]byte) [sha256.Size]byte {
	tmp := append(oraclePrivKey, sdk.Uint64ToBigEndian(dealID)...)
	return sha256.Sum256(append(tmp, dataHash[:]...))
}
