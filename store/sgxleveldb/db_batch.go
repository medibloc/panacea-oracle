package sgxleveldb

import (
	"github.com/medibloc/panacea-oracle/sgx"
	log "github.com/sirupsen/logrus"
	tmdb "github.com/tendermint/tm-db"
)

type sgxLevelDBBatch struct {
	tmdb.Batch
}

func (sbatch *sgxLevelDBBatch) Set(key, value []byte) error {
	log.Debug("sealing before writing to leveldb in batch")
	sealValue, err := sgx.Seal(value)
	if err != nil {
		return err
	}
	return sbatch.Batch.Set(key, sealValue)
}
