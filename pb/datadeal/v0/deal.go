package v0

import (
	"fmt"

	"github.com/medibloc/panacea-oracle/panacea"
)

func (r *ValidateDataRequest) ValidateBasic() error {
	if _, err := panacea.GetAccAddressFromBech32(r.ProviderAddress); err != nil {
		return fmt.Errorf("invalid provider address: %w", err)
	}

	if len(r.EncryptedData) == 0 {
		return fmt.Errorf("encrypted data is empty in request")
	}

	if len(r.DataHash) == 0 {
		return fmt.Errorf("data hash is empty in request")
	}

	return nil
}
