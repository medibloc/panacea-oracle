package service

import (
	"crypto/sha256"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GetCombinedKey(oraclePrivKey []byte, dealID uint64, dataHash []byte) []byte {
	hash := sha256.New()
	hash.Write(oraclePrivKey)
	hash.Write(sdk.Uint64ToBigEndian(dealID))
	hash.Write(dataHash)
	return hash.Sum(nil)
}
