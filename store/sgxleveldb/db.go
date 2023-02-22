// Package leveldb implements a light client level db for panacea-doracle.
// It does include Set & Get functions that are sealed & unsealed in the sgx environment.

package sgxleveldb

import (
	"fmt"

	"github.com/medibloc/panacea-oracle/sgx"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb/opt"
	tmdb "github.com/tendermint/tm-db"
)

type SgxLevelDB struct {
	sgx sgx.Sgx
	*tmdb.GoLevelDB
}

func NewSgxLevelDB(name string, dir string, sgx sgx.Sgx) (*SgxLevelDB, error) {
	return NewSgxLevelDBWithOpts(name, dir, sgx, nil)
}

func NewSgxLevelDBWithOpts(name string, dir string, sgx sgx.Sgx, o *opt.Options) (*SgxLevelDB, error) {
	goLevelDB, err := tmdb.NewGoLevelDBWithOpts(name, dir, o)
	if err != nil {
		return nil, fmt.Errorf("failed to NewGoLevelDBWithOpts: %w", err)
	}

	return &SgxLevelDB{sgx, goLevelDB}, nil
}

func (sdb *SgxLevelDB) Set(key, value []byte) error {
	log.Debug("sealing before writing to leveldb")
	sealValue, err := sdb.sgx.Seal(value)
	if err != nil {
		return err
	}
	return sdb.GoLevelDB.Set(key, sealValue)
}

func (sdb *SgxLevelDB) Get(key []byte) ([]byte, error) {
	val, err := sdb.GoLevelDB.Get(key)
	if err != nil {
		return nil, err
	} else if val == nil {
		return nil, nil
	}

	log.Debug("unsealing after reading from leveldb")
	unsealedVal, err := sdb.sgx.Unseal(val)
	if err != nil {
		return nil, fmt.Errorf("failed to unseal value from leveldb: %w", err)
	}

	return unsealedVal, nil
}

func (sdb *SgxLevelDB) NewBatch() tmdb.Batch {
	batch := sdb.GoLevelDB.NewBatch()
	return &sgxLevelDBBatch{sdb.sgx, batch}
}
