package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-oracle/config"
	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/sgx"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	tos "github.com/tendermint/tendermint/libs/os"
)

func getOracleKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-oracle-key",
		Short: "Get a shared oracle private key",
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := loadConfigFromHome(cmd)
			if err != nil {
				return err
			}

			ctx := context.Background()

			nodePrivKeyPath := conf.AbsNodePrivKeyPath()
			if !tos.FileExists(nodePrivKeyPath) {
				return errors.New("no node_priv_key.sealed file")
			}

			nodePrivKeyBz, err := sgx.UnsealFromFile(nodePrivKeyPath)
			if err != nil {
				return fmt.Errorf("failed to unseal node_priv_key.sealed file: %w", err)
			}
			nodePrivKey, nodePubKey := crypto.PrivKeyFromBytes(nodePrivKeyBz)

			// get oracle account from mnemonic.
			oracleAccount, err := panacea.NewOracleAccount(conf.OracleMnemonic, conf.OracleAccNum, conf.OracleAccIndex)
			if err != nil {
				return fmt.Errorf("failed to get oracle account from mnemonic: %w", err)
			}

			// get OracleRegistration from Panacea
			queryClient, err := panacea.LoadVerifiedQueryClient(ctx, conf)
			if err != nil {
				return fmt.Errorf("failed to get queryClient: %w", err)
			}
			defer queryClient.Close()

			// get unique ID
			selfEnclaveInfo, err := sgx.GetSelfEnclaveInfo()
			if err != nil {
				return fmt.Errorf("failed to get self enclave info: %w", err)
			}
			uniqueID := selfEnclaveInfo.UniqueIDHex()

			oracleRegistration, err := queryClient.GetOracleRegistration(oracleAccount.GetAddress(), uniqueID)
			if err != nil {
				return fmt.Errorf("failed to get oracle registration from Panacea: %w", err)
			}

			if oracleRegistration.EncryptedOraclePrivKey == nil {
				return fmt.Errorf("failed to get encrypted oracle private key")
			}

			// check if the same node key is used for oracle registration
			if !bytes.Equal(oracleRegistration.NodePubKey, nodePubKey.SerializeCompressed()) {
				return errors.New("the existing node key is different from the one used in oracle registration. if you want to re-request RegisterOracle, delete the existing node_priv_key.sealed file and rerun register-oracle cmd")
			}

			oraclePublicKey, err := queryClient.GetOracleParamsPublicKey()
			if err != nil {
				return err
			}

			return getOraclePrivKey(conf, oracleRegistration, nodePrivKey, oraclePublicKey)
		},
	}

	return cmd
}

func getOraclePrivKey(conf *config.Config, oracleRegistration *oracletypes.OracleRegistration, nodePrivKey *btcec.PrivateKey, oraclePubKey *btcec.PublicKey) error {
	oraclePrivKeyPath := conf.AbsOraclePrivKeyPath()
	if tos.FileExists(oraclePrivKeyPath) {
		return errors.New("the oracle private key already exists")
	}

	shareKey := crypto.DeriveSharedKey(nodePrivKey, oraclePubKey, crypto.KDFSHA256)

	oraclePrivKey, err := crypto.DecryptWithAES256(shareKey, oracleRegistration.EncryptedOraclePrivKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt the encrypted oracle private key: %w", err)
	}

	if err := sgx.SealToFile(oraclePrivKey, oraclePrivKeyPath); err != nil {
		return fmt.Errorf("failed to seal to file: %w", err)
	}

	log.Info("oracle private key is retrieved successfully")
	return nil
}
