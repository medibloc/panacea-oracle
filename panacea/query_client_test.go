package panacea_test

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/go-bip39"
	"github.com/medibloc/panacea-oracle/integration/suite"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tm-db"
)

type queryClientTestSuite struct {
	suite.TestSuite

	chainID           string
	validatorMnemonic string
}

func TestQueryClient(t *testing.T) {
	initScriptPath, err := filepath.Abs("testdata/panacea-core-init.sh")
	require.NoError(t, err)

	chainID := "testing"
	entropy, err := bip39.NewEntropy(256)
	require.NoError(t, err)
	validatorMnemonic, err := bip39.NewMnemonic(entropy)
	require.NoError(t, err)

	suite.Run(t, &queryClientTestSuite{
		suite.NewTestSuite(
			initScriptPath,
			[]string{
				fmt.Sprintf("CHAIN_ID=%s", chainID),
				fmt.Sprintf("MNEMONIC=%s", validatorMnemonic),
			},
		),
		chainID,
		validatorMnemonic,
	})
}

func (suite *queryClientTestSuite) TestGetAccount() {
	trustedBlockInfo, conf := suite.Prepare(suite.chainID)

	ctx := context.Background()
	queryClient, err := panacea.NewVerifiedQueryClientWithDB(ctx, conf, trustedBlockInfo, dbm.NewMemDB())
	require.NoError(suite.T(), err)
	defer queryClient.Close()

	var wg sync.WaitGroup
	address := panacea.GetAddressFromMnemonic(suite.validatorMnemonic, 0, 0)

	for i := 0; i < 10; i++ { // to check if queryClient is goroutine-safe
		wg.Add(1)

		go func() {
			defer wg.Done()

			acc, err := queryClient.GetAccount(ctx, address)
			require.NoError(suite.T(), err)

			address, err := bech32.ConvertAndEncode("panacea", acc.GetPubKey().Address().Bytes())
			require.NoError(suite.T(), err)
			require.Equal(suite.T(), address, address)
		}()
	}

	wg.Wait()
}

func (suite *queryClientTestSuite) TestLoadQueryClient() {
	trustedBlockInfo, conf := suite.Prepare(suite.chainID)

	db := dbm.NewMemDB()

	ctx := context.Background()

	queryClient, err := panacea.NewVerifiedQueryClientWithDB(ctx, conf, trustedBlockInfo, db)
	require.NoError(suite.T(), err)

	lastTrustedHeight, err := queryClient.GetLastBlockHeight(ctx)
	require.NoError(suite.T(), err)
	require.GreaterOrEqual(suite.T(), lastTrustedHeight, trustedBlockInfo.TrustedBlockHeight)

	err = queryClient.Close() // here, memdb is not closed because MemDB.Close() is actually empty
	require.NoError(suite.T(), err)

	// try to load query client, instead of creating it
	queryClient, err = panacea.NewVerifiedQueryClientWithDB(context.Background(), conf, nil, db)
	require.NoError(suite.T(), err)

	lastTrustedHeight2, err := queryClient.GetLastBlockHeight(ctx)
	require.NoError(suite.T(), err)
	require.GreaterOrEqual(suite.T(), lastTrustedHeight2, lastTrustedHeight)
}

func (suite *queryClientTestSuite) TestGetOracleUpgradeInfoEmptyValue() {
	trustedBlockInfo, conf := suite.Prepare(suite.chainID)

	ctx := context.Background()
	queryClient, err := panacea.NewVerifiedQueryClientWithDB(ctx, conf, trustedBlockInfo, dbm.NewMemDB())
	require.NoError(suite.T(), err)
	defer queryClient.Close()

	upgradeInfo, err := queryClient.GetOracleUpgradeInfo(ctx)
	require.Nil(suite.T(), upgradeInfo)
	require.ErrorIs(suite.T(), err, panacea.ErrEmptyValue)
}