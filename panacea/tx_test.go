package panacea_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-oracle/config"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/types/test_utils/mocks"
	"github.com/stretchr/testify/require"
)

func TestGenerateTxBytes(t *testing.T) {
	conf := config.DefaultConfig()
	privKey := secp256k1.GenPrivKey()
	// it is a simple msg.
	commissionRate := sdk.NewDecWithPrec(1, 1)
	msg := &oracletypes.MsgUpdateOracleInfo{
		OracleAddress:        "oracle_address",
		Endpoint:             "end_point",
		OracleCommissionRate: &commissionRate,
	}

	signerAccount := mocks.NewMockAccount(privKey.PubKey())
	chainID := "chainID"
	builder := panacea.NewTxBuilder(
		mocks.MockGrpcClient{
			Account: signerAccount,
			ChainID: chainID,
		})
	txBodyBz, err := builder.GenerateTxBytes(privKey, conf, msg)
	require.NoError(t, err)

	var txRaw tx.TxRaw
	err = txRaw.Unmarshal(txBodyBz)
	require.NoError(t, err)

	var txBody tx.TxBody
	err = txBody.Unmarshal(txRaw.BodyBytes)
	require.NoError(t, err)

	var authInfo tx.AuthInfo
	err = authInfo.Unmarshal(txRaw.AuthInfoBytes)
	require.NoError(t, err)

	defaultFeeAmount, err := sdk.ParseCoinsNormalized(conf.Panacea.DefaultFeeAmount)
	require.NoError(t, err)
	require.Equal(t, defaultFeeAmount, authInfo.GetFee().GetAmount())

	authInfoBz, err := authInfo.Marshal()
	require.NoError(t, err)

	signDoc := tx.SignDoc{
		BodyBytes:     txRaw.GetBodyBytes(),
		AuthInfoBytes: authInfoBz,
		ChainId:       chainID,
		AccountNumber: signerAccount.AccountNumber,
	}

	signDocBz, err := signDoc.Marshal()
	require.NoError(t, err)
	require.True(t, privKey.PubKey().VerifySignature(signDocBz, txRaw.Signatures[0]))
}
