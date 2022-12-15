package cmd

import (
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-oracle/config"
	"github.com/medibloc/panacea-oracle/crypto"
	oracleservice "github.com/medibloc/panacea-oracle/service/oracle"
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

			svc, err := oracleservice.New(conf)
			if err != nil {
				return fmt.Errorf("failed to create service: %w", err)
			}
			defer svc.Close()

			if err := svc.GetAndStoreOraclePrivKey(); err != nil {
				return err
			}

			return nil
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

	oraclePrivKey, err := crypto.Decrypt(shareKey, nil, oracleRegistration.EncryptedOraclePrivKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt the encrypted oracle private key: %w", err)
	}

	if err := sgx.SealToFile(oraclePrivKey, oraclePrivKeyPath); err != nil {
		return fmt.Errorf("failed to seal to file: %w", err)
	}

	log.Info("oracle private key is retrieved successfully")
	return nil
}
