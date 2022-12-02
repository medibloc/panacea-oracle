package sgx

import (
	"encoding/hex"
	"fmt"
	"github.com/edgelesssys/ego/enclave"
)

const dummyData = "dummy-data"

type EnclaveInfo struct {
	ProductID []byte
	SignerID  []byte
	UniqueID  []byte
}

func NewEnclaveInfo(productID, signerID, uniqueID []byte) *EnclaveInfo {
	return &EnclaveInfo{
		ProductID: productID,
		SignerID:  signerID,
		UniqueID:  uniqueID,
	}
}

// GetSelfEnclaveInfo sets EnclaveInfo from self-generated remote report
func GetSelfEnclaveInfo() (*EnclaveInfo, error) {
	// generate self-remote-report and get product ID, signer ID, and unique ID
	reportBz, err := GenerateRemoteReport([]byte(dummyData))
	if err != nil {
		return nil, fmt.Errorf("failed to generate self-report: %w", err)
	}

	report, err := enclave.VerifyRemoteReport(reportBz)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve self-report: %w", err)
	}

	return NewEnclaveInfo(report.ProductID, report.SignerID, report.UniqueID), nil
}

func (e EnclaveInfo) UniqueIDHex() string {
	return hex.EncodeToString(e.UniqueID)
}
