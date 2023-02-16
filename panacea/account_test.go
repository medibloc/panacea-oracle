package panacea_test

import (
	"strings"
	"testing"

	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/stretchr/testify/require"
)

func TestNewOracleAccount(t *testing.T) {
	mnemonic, err := crypto.NewMnemonic()
	require.NoError(t, err)

	oracleAcc, err := panacea.NewOracleAccount(mnemonic, 0, 0)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(oracleAcc.GetAddress(), "panacea1"))
	require.Equal(t, oracleAcc.GetPubKey().Address().Bytes(), oracleAcc.AccAddressFromBech32().Bytes())
}
