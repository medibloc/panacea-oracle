package ipfs

import (
	"bytes"

	shell "github.com/ipfs/go-ipfs-api"
	log "github.com/sirupsen/logrus"
)

type IPFS struct {
	sh *shell.Shell
}

// NewIPFS generates an IPFS node with IPFS url.
func NewIPFS(url string) *IPFS {
	newShell := shell.NewShell(url)

	if !newShell.IsUp() {
		log.Errorf("IPFS is not connected")
	} else {
		log.Info("successfully connect to IPFS node")
	}

	return &IPFS{
		sh: newShell,
	}
}

// Add method adds a data and returns a CID.
func (i *IPFS) Add(data []byte) (string, error) {
	reader := bytes.NewReader(data)

	cid, err := i.sh.Add(reader)
	if err != nil {
		return "", err
	}

	return cid, nil
}

// Get method gets a data and returns a bytes of Deal.
func (i *IPFS) Get(cid string) ([]byte, error) {
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
