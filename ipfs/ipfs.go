package ipfs

import (
	"bytes"
	"fmt"

	shell "github.com/ipfs/go-ipfs-api"
)

type IPFS struct {
	url string
}

// NewIPFS generates an IPFS node with IPFS url.
func NewIPFS(url string) (*IPFS, error) {
	// As with the previous specification, when an IPFS object is initially created, it checks to see if IPFS can connect well.
	if !shell.NewShell(url).IsUp() {
		return nil, fmt.Errorf("IPFS is not connected")
	}

	return &IPFS{
		url: url,
	}, nil
}

func (i *IPFS) createShell() *shell.Shell {
	return shell.NewShell(i.url)
}

// Add method adds a data and returns a CID.
func (i *IPFS) Add(data []byte) (string, error) {
	reader := bytes.NewReader(data)

	cid, err := i.createShell().Add(reader)
	if err != nil {
		return "", err
	}

	return cid, nil
}

// Get method gets a data and returns a bytes of Deal.
func (i *IPFS) Get(cid string) ([]byte, error) {
	data, err := i.createShell().Cat(cid)
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
