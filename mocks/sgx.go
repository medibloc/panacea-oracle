package mocks

import (
	"io/fs"
	"os"

	"github.com/medibloc/panacea-oracle/sgx"
)

type MockSGX struct {
	RemoteReport            []byte
	SelfEnclaveInfo         *sgx.EnclaveInfo
	VerifyRemoteReportError error
}

var _ sgx.Sgx = &MockSGX{}

func (m MockSGX) GenerateRemoteReport(data []byte) ([]byte, error) {
	return m.RemoteReport, nil
}

func (m MockSGX) GenerateSelfEnclaveInfo() (*sgx.EnclaveInfo, error) {
	return m.SelfEnclaveInfo, nil
}

func (m MockSGX) VerifyRemoteReport(reportBytes, expectedData []byte, expectedUniqueID []byte) error {
	return m.VerifyRemoteReportError
}

func (m MockSGX) SealToFile(data []byte, filePath string) error {
	return os.WriteFile(filePath, data, fs.ModePerm)
}

func (m MockSGX) UnsealFromFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

func (m MockSGX) Seal(data []byte) ([]byte, error) {
	return data, nil
}

func (m MockSGX) Unseal(data []byte) ([]byte, error) {
	return data, nil
}
