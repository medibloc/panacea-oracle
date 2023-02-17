package ipfs

import (
	"bytes"
	"fmt"

	shell "github.com/ipfs/go-ipfs-api"
	log "github.com/sirupsen/logrus"
)

type IPFS interface {
	Add(data []byte) (string, error)

	Get(cid string) ([]byte, error)
}

var _ IPFS = &ipfs{}

type ipfs struct {
	sh *shell.Shell
}

// NewIPFS generates an IPFS node with IPFS url.
func NewIPFS(url string) (IPFS, error) {
	newShell := shell.NewShell(url)

	if !newShell.IsUp() {
		return nil, fmt.Errorf("IPFS is not connected")
	}

	log.Info("successfully connect to IPFS node")

	return &ipfs{
		sh: newShell,
	}, nil
}

// Add method adds a data and returns a CID.
func (i *ipfs) Add(data []byte) (string, error) {
	reader := bytes.NewReader(data)

	cid, err := i.sh.Add(reader)
	if err != nil {
		return "failed to add data to IPFS", err
	}

	return cid, nil
}

// Get method gets a data and returns a bytes of Deal.
func (i *ipfs) Get(cid string) ([]byte, error) {
	data, err := i.sh.Cat(cid)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
