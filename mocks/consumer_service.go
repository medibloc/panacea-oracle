package mocks

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/btcsuite/btcd/btcec"
	"github.com/medibloc/panacea-oracle/consumer_service"
)

type MockConsumerService struct {
	oraclePrivKey *btcec.PrivateKey
}

var (
	_ consumer_service.FileStorage = &MockConsumerService{}
)

func (u MockConsumerService) Add(tempDir string, dealID uint64, dataHash string, data []byte) error {
	if err := os.MkdirAll(filepath.Join(tempDir, strconv.FormatUint(dealID, 10)), fs.ModePerm); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(tempDir, strconv.FormatUint(dealID, 10), dataHash), data, fs.ModePerm); err != nil {
		return err
	}

	return nil
}

func (u MockConsumerService) Get(tempDir string, dealID uint64, dataHash string) ([]byte, error) {
	return os.ReadFile(filepath.Join(tempDir, strconv.FormatUint(dealID, 10), dataHash))
}
