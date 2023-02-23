package sgx

import (
	"encoding/hex"
)

const dummyData = "dummy-data"

type EnclaveInfo struct {
	ProductID []byte
	UniqueID  []byte
}

func NewEnclaveInfo(productID, uniqueID []byte) *EnclaveInfo {
	return &EnclaveInfo{
		ProductID: productID,
		UniqueID:  uniqueID,
	}
}

func (e EnclaveInfo) UniqueIDHex() string {
	return hex.EncodeToString(e.UniqueID)
}
