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
