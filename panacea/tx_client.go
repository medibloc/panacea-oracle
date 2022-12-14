package panacea

import (
	"fmt"

	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/medibloc/panacea-oracle/config"
)

type TxClient struct {
	defaultFeeAmount string
	defaultGasLimit  uint64

	oraclePrivKey cryptotypes.PrivKey

	grpcClient  *GRPCClient
	queryClient QueryClient
}

func NewTxClient(conf *config.Config, grpcClient *GRPCClient, queryClient QueryClient) (*TxClient, error) {
	oracleAccount, err := NewOracleAccount(conf.OracleMnemonic, conf.OracleAccNum, conf.OracleAccIndex)
	if err != nil {
		return nil, err
	}

	return &TxClient{
		conf.Panacea.DefaultFeeAmount,
		conf.Panacea.DefaultGasLimit,
		oracleAccount.privKey,
		grpcClient,
		queryClient,
	}, nil
}

func (tc *TxClient) BroadcastTx(msg sdk.Msg) (int64, string, error) {
	defaultFeeAmount, _ := sdk.ParseCoinsNormalized(tc.defaultFeeAmount)

	txBytes, err := tc.GenerateSignedTxBytes(
		tc.oraclePrivKey,
		tc.defaultGasLimit,
		defaultFeeAmount,
		msg,
	)
	if err != nil {
		return 0, "", fmt.Errorf("failed to generate signed Tx bytes: %w", err)
	}

	resp, err := tc.grpcClient.BroadcastTx(txBytes)

	if err != nil {
		return 0, "", fmt.Errorf("broadcast transaction failed. txBytes(%v)", txBytes)
	}

	if resp.TxResponse.Code != 0 {
		return 0, "", fmt.Errorf("transaction failed: %v", resp.TxResponse.RawLog)
	}

	return resp.TxResponse.Height, resp.TxResponse.TxHash, nil
}

// GenerateTxBytes generates transaction byte array.
func (tc *TxClient) GenerateTxBytes(privKey cryptotypes.PrivKey, conf *config.Config, msg ...sdk.Msg) ([]byte, error) {
	defaultFeeAmount, err := sdk.ParseCoinsNormalized(conf.Panacea.DefaultFeeAmount)
	if err != nil {
		return nil, err
	}
	txBytes, err := tc.GenerateSignedTxBytes(privKey, conf.Panacea.DefaultGasLimit, defaultFeeAmount, msg...)
	if err != nil {
		return nil, err
	}

	return txBytes, nil
}

// GenerateSignedTxBytes signs msgs using the private key and returns the signed Tx message in form of byte array.
func (tc *TxClient) GenerateSignedTxBytes(
	privateKey cryptotypes.PrivKey,
	gasLimit uint64,
	feeAmount sdk.Coins,
	msg ...sdk.Msg,
) ([]byte, error) {
	txConfig := authtx.NewTxConfig(tc.queryClient.GetCdc(), []signing.SignMode{signing.SignMode_SIGN_MODE_DIRECT})
	txBuilder := txConfig.NewTxBuilder()
	txBuilder.SetGasLimit(gasLimit)
	txBuilder.SetFeeAmount(feeAmount)

	if err := txBuilder.SetMsgs(msg...); err != nil {
		return nil, err
	}

	signerAddress, err := bech32.ConvertAndEncode(prefix, privateKey.PubKey().Address().Bytes())
	if err != nil {
		return nil, err
	}

	signerAccount, err := tc.queryClient.GetAccount(signerAddress)
	if err != nil {
		return nil, err
	}

	sigV2 := signing.SignatureV2{
		PubKey: privateKey.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: nil,
		},
		Sequence: signerAccount.GetSequence(),
	}

	if err := txBuilder.SetSignatures(sigV2); err != nil {
		return nil, err
	}

	signerData := authsigning.SignerData{
		ChainID:       tc.queryClient.GetChainID(),
		AccountNumber: signerAccount.GetAccountNumber(),
		Sequence:      signerAccount.GetSequence(),
	}

	sigV2, err = clienttx.SignWithPrivKey(
		signing.SignMode_SIGN_MODE_DIRECT,
		signerData,
		txBuilder,
		privateKey,
		txConfig,
		signerAccount.GetSequence(),
	)
	if err != nil {
		return nil, err
	}

	if err := txBuilder.SetSignatures(sigV2); err != nil {
		return nil, err
	}

	return txConfig.TxEncoder()(txBuilder.GetTx())
}
