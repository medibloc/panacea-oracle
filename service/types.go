package service

import (
	"encoding/base64"
	"fmt"

	"github.com/medibloc/panacea-oracle/panacea"
)

type ValidateDataReq struct {
	ProviderAddress     string `json:"provider_address"`
	EncryptedDataBase64 string `json:"encrypted_data_base64"`
	DataHash            string `json:"data_hash"`
}

func (r *ValidateDataReq) ValidateBasic() error {
	if _, err := panacea.GetAccAddressFromBech32(r.ProviderAddress); err != nil {
		return fmt.Errorf("invalid provider address: %w", err)
	}

	if _, err := base64.StdEncoding.DecodeString(r.EncryptedDataBase64); err != nil {
		return fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	if len(r.DataHash) == 0 {
		return fmt.Errorf("data hash is empty in request")
	}

	return nil
}

type secretKeyResponse struct {
	EncryptedSecretKey []byte `json:"encrypted_secret_key"`
}

// statusResponse is a response type for GET /v0/status.
// Not using existing structs directly, such as sgx.EnclaveInfo, since they may contain sensitive values in the future.
// TODO: feel free to add more values, such as version
type statusResponse struct {
	OracleAccountAddress string            `json:"oracle_account_address"`
	API                  statusAPI         `json:"api"`
	EnclaveInfo          statusEnclaveInfo `json:"enclave_info"`
}

type statusAPI struct {
	ListenAddr string `json:"listen_addr"`
}

type statusEnclaveInfo struct {
	ProductIDBase64 string `json:"product_id_base64"`
	UniqueIDBase64  string `json:"unique_id_base64"`
}
