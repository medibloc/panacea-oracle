package mocks

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/medibloc/panacea-oracle/ipfs"
)

// MockIPFS write and read files from local storage
type MockIPFS struct {
}

var (
	_ ipfs.IPFS = &MockIPFS{}

	defaultPath = "ipfs"
)

func (u MockIPFS) Add(data []byte) (string, error) {
	os.MkdirAll(defaultPath, fs.ModePerm)

	ran := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, ran); err != nil {
		return "", err
	}
	cid := hex.EncodeToString(ran)

	if err := os.WriteFile(filepath.Join(defaultPath, cid), data, fs.ModePerm); err != nil {
		return "", err
	}

	return cid, nil
}

func (u MockIPFS) Get(cid string) ([]byte, error) {
	return os.ReadFile(filepath.Join(defaultPath, cid))
}

func RemoveMockIPFSData() {
	if info, _ := os.Stat(defaultPath); info != nil {
		os.RemoveAll(defaultPath)
	}
}
