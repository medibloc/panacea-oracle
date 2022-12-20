package panacea

import (
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/ibc-go/v2/modules/core/23-commitment/types"
	"github.com/stretchr/testify/require"
)

func TestFoo(t *testing.T) {
	addrStr := "panacea12q8pmd3kykn8p6hnzc5chxj4l0479l2uxgfr7m"
	addr, err := GetAccAddressFromBech32(addrStr)
	require.NoError(t, err)

	key := authtypes.AddressStoreKey(addr)

	merklePath := types.NewMerklePath(authtypes.StoreKey, string(key))

	_, err = merklePath.GetKey(uint64(len(merklePath.KeyPath) - 1 - 0))
	require.NoError(t, err)
}
