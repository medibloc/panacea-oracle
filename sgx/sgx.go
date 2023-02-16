package sgx

import (
	"bytes"
	"fmt"
	"os"

	"github.com/edgelesssys/ego/ecrypto"
	"github.com/edgelesssys/ego/enclave"
	log "github.com/sirupsen/logrus"
)

type Sgx interface {
	GenerateRemoteReport(data []byte) ([]byte, error)

	GenerateSelfEnclaveInfo() (*EnclaveInfo, error)

	VerifyRemoteReport(reportBytes, expectedData []byte, expectedUniqueID []byte) error

	SealToFile(data []byte, filePath string) error

	UnsealFromFile(filePath string) ([]byte, error)

	Seal(data []byte) ([]byte, error)

	Unseal(data []byte) ([]byte, error)
}

var _ Sgx = oracleSgx{}

type oracleSgx struct{}

func NewOracleSgx() Sgx {
	return &oracleSgx{}
}

func (s oracleSgx) GenerateRemoteReport(data []byte) ([]byte, error) {
	return enclave.GetRemoteReport(data)
}

// GenerateSelfEnclaveInfo sets EnclaveInfo from self-generated remote report
func (s oracleSgx) GenerateSelfEnclaveInfo() (*EnclaveInfo, error) {
	// generate self-remote-report and get product ID, signer ID, and unique ID
	reportBz, err := s.GenerateRemoteReport([]byte(dummyData))
	if err != nil {
		return nil, fmt.Errorf("failed to generate self-report: %w", err)
	}

	report, err := enclave.VerifyRemoteReport(reportBz)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve self-report: %w", err)
	}

	return NewEnclaveInfo(report.ProductID, report.UniqueID), nil
}

func (s oracleSgx) VerifyRemoteReport(reportBytes, expectedData []byte, expectedUniqueID []byte) error {
	report, err := enclave.VerifyRemoteReport(reportBytes)
	if err != nil {
		return err
	}

	if report.SecurityVersion < PromisedMinSecurityVersion {
		return fmt.Errorf("invalid security version in the report")
	}
	if !bytes.Equal(report.UniqueID, expectedUniqueID) {
		return fmt.Errorf("invalid unique ID in the report")
	}
	if !bytes.Equal(report.Data[:len(expectedData)], expectedData) {
		return fmt.Errorf("invalid data in the report")
	}

	return nil
}

func (s oracleSgx) SealToFile(data []byte, filePath string) error {
	sealedData, err := s.Seal(data)
	if err != nil {
		return fmt.Errorf("failed to seal oracle private key: %w", err)
	}

	if err := os.WriteFile(filePath, sealedData, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", filePath, err)
	}
	log.Infof("%s is sealed and written successfully", filePath)

	return nil
}

func (s oracleSgx) UnsealFromFile(filePath string) ([]byte, error) {
	sealed, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	key, err := s.Unseal(sealed)
	if err != nil {
		return nil, fmt.Errorf("failed to unseal oracle key: %w", err)
	}

	return key, nil
}

func (s oracleSgx) Seal(data []byte) ([]byte, error) {
	return ecrypto.SealWithUniqueKey(data, nil)
}

func (s oracleSgx) Unseal(data []byte) ([]byte, error) {
	return ecrypto.Unseal(data, nil)
}
