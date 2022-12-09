package sgx

import (
	"fmt"
	"os"

	"github.com/edgelesssys/ego/ecrypto"
	log "github.com/sirupsen/logrus"
)

// SealToFile seals the data with unique ID and stores it to file.
func SealToFile(data []byte, filePath string) error {
	sealedData, err := Seal(data)
	if err != nil {
		return fmt.Errorf("failed to seal oracle private key: %w", err)
	}

	if err := os.WriteFile(filePath, sealedData, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", filePath, err)
	}
	log.Infof("%s is sealed and written successfully", filePath)

	return nil
}

func UnsealFromFile(filePath string) ([]byte, error) {
	sealed, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	key, err := Unseal(sealed)
	if err != nil {
		return nil, fmt.Errorf("failed to unseal oracle key: %w", err)
	}

	return key, nil
}

// Seal returns data sealed with unique ID in SGX-enabled environments
func Seal(data []byte) ([]byte, error) {
	return ecrypto.SealWithUniqueKey(data, nil)
}

// Unseal returns data unsealed with unique ID in SGX-enabled environments
func Unseal(data []byte) ([]byte, error) {
	return ecrypto.Unseal(data, nil)
}
