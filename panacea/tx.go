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

type TxBuilder struct {
	client QueryClient
}

func NewTxBuilder(client QueryClient) *TxBuilder {
	return &TxBuilder{
		client: client,
	}
}

// GenerateTxBytes generates transaction byte array.
func (tb TxBuilder) GenerateTxBytes(privKey cryptotypes.PrivKey, conf *config.Config, msg ...sdk.Msg) ([]byte, error) {
	defaultFeeAmount, err := sdk.ParseCoinsNormalized(conf.Panacea.DefaultFeeAmount)
	if err != nil {
		return nil, err
	}
	txBytes, err := tb.GenerateSignedTxBytes(privKey, conf.Panacea.DefaultGasLimit, defaultFeeAmount, msg...)
	if err != nil {
		return nil, err
	}

	return txBytes, nil
}

// GenerateSignedTxBytes signs msgs using the private key and returns the signed Tx message in form of byte array.
func (tb TxBuilder) GenerateSignedTxBytes(
	privateKey cryptotypes.PrivKey,
	gasLimit uint64,
	feeAmount sdk.Coins,
	msg ...sdk.Msg,
) ([]byte, error) {
	txConfig := authtx.NewTxConfig(tb.client.GetCdc(), []signing.SignMode{signing.SignMode_SIGN_MODE_DIRECT})
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

	signerAccount, err := tb.client.GetAccount(signerAddress)
	if err != nil {
		return nil, fmt.Errorf("can not get signer account from address(%s): %w", signerAddress, err)
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
		ChainID:       tb.client.GetChainID(),
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
