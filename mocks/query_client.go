package mocks

import (
	"context"
	didtypes "github.com/medibloc/panacea-core/v2/x/did/types"

	"github.com/btcsuite/btcd/btcec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	datadealtypes "github.com/medibloc/panacea-core/v2/x/datadeal/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-oracle/panacea"
	tmtypes "github.com/tendermint/tendermint/types"
)

// MockQueryClient is a very simple mock structure.
// It is implemented to return the value as it is declared in this mock structure.
type MockQueryClient struct {
	Account                     authtypes.AccountI
	AccountError                error
	OracleRegistration          *oracletypes.OracleRegistration
	LightBlock                  *tmtypes.LightBlock
	LastBlockHeight             int64
	OraclePubKey                *btcec.PublicKey
	Deal                        *datadealtypes.Deal
	Consent                     *datadealtypes.Consent
	Oracle                      *oracletypes.Oracle
	OracleUpgrade               *oracletypes.OracleUpgrade
	OracleUpgradeInfo           *oracletypes.OracleUpgradeInfo
	VerifyTrustedBlockInfoError error
	DidDocWithSeq               *didtypes.DIDDocumentWithSeq
}

var _ panacea.QueryClient = &MockQueryClient{}

func (q MockQueryClient) Close() error {
	return nil
}

func (q MockQueryClient) GetAccount(ctx context.Context, s string) (authtypes.AccountI, error) {
	return q.Account, q.AccountError
}

func (q MockQueryClient) GetDID(_ context.Context, _ string) (*didtypes.DIDDocumentWithSeq, error) {
	return q.DidDocWithSeq, nil
}

func (q MockQueryClient) GetOracleRegistration(ctx context.Context, s string, s2 string) (*oracletypes.OracleRegistration, error) {
	return q.OracleRegistration, nil
}

func (q MockQueryClient) GetLightBlock(height int64) (*tmtypes.LightBlock, error) {
	return q.LightBlock, nil
}

func (q MockQueryClient) GetOracleParamsPublicKey(ctx context.Context) (*btcec.PublicKey, error) {
	return q.OraclePubKey, nil
}

func (q MockQueryClient) GetDeal(ctx context.Context, u2 uint64) (*datadealtypes.Deal, error) {
	return q.Deal, nil
}

func (q MockQueryClient) GetConsent(ctx context.Context, u2 uint64, s string) (*datadealtypes.Consent, error) {
	return q.Consent, nil
}

func (q MockQueryClient) GetLastBlockHeight(ctx context.Context) (int64, error) {
	return q.LastBlockHeight, nil
}

func (q MockQueryClient) GetOracleUpgrade(ctx context.Context, s string, s2 string) (*oracletypes.OracleUpgrade, error) {
	return q.OracleUpgrade, nil
}

func (q MockQueryClient) GetOracleUpgradeInfo(ctx context.Context) (*oracletypes.OracleUpgradeInfo, error) {
	return q.OracleUpgradeInfo, nil
}

func (q MockQueryClient) GetOracle(ctx context.Context, s string) (*oracletypes.Oracle, error) {
	return q.Oracle, nil
}

func (q MockQueryClient) VerifyTrustedBlockInfo(i int64, bytes []byte) error {
	return q.VerifyTrustedBlockInfoError
}
