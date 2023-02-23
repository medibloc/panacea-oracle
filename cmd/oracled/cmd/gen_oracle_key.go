package cmd

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/medibloc/panacea-oracle/client/flags"
	"github.com/medibloc/panacea-oracle/panacea"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/medibloc/panacea-oracle/sgx"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	tos "github.com/tendermint/tendermint/libs/os"
)

// OraclePubKeyInfo is a struct to store oracle public key and its remote report
type OraclePubKeyInfo struct {
	PublicKeyBase64    string `json:"public_key_base64"`
	RemoteReportBase64 string `json:"remote_report_base64"`
}

func genOracleKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen-oracle-key",
		Short: "Generate oracle key",
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := loadConfigFromHome(cmd)
			if err != nil {
				return err
			}

			// If there is the existing oracle key, double-check for generating a new oracle key
			oraclePrivKeyPath := conf.AbsOraclePrivKeyPath()
			if tos.FileExists(oraclePrivKeyPath) {
				buf := bufio.NewReader(os.Stdin)
				ok, err := input.GetConfirmation("This can replace the existing oracle-priv-key.sealed file.\nAre you sure to make a new oracle key?", buf, os.Stderr)

				if err != nil || !ok {
					log.Printf("Oracle key generation is canceled.")
					return err
				}
			}

			// get trusted block information
			trustedBlockInfo, err := getTrustedBlockInfo(cmd)
			if err != nil {
				return fmt.Errorf("failed to get trusted block info: %w", err)
			}

			// generate a new oracle key
			oraclePrivKey, err := crypto.NewPrivKey()
			if err != nil {
				log.Errorf("failed to generate oracle key: %v", err)
				return err
			}

			sgx := sgx.NewOracleSGX()

			// seal and store oracle private key
			if err := sgx.SealToFile(oraclePrivKey.Serialize(), oraclePrivKeyPath); err != nil {
				log.Errorf("failed to write %s: %v", oraclePrivKeyPath, err)
				return err
			}

			// generate oracle key remote report
			oraclePubKey := oraclePrivKey.PubKey().SerializeCompressed()
			oraclePubKeyHash := sha256.Sum256(oraclePubKey)
			oracleKeyRemoteReport, err := sgx.GenerateRemoteReport(oraclePubKeyHash[:])
			if err != nil {
				log.Errorf("failed to generate remote report of oracle key: %v", err)
				return err
			}

			// store oracle pub key and its remote report to a file
			if err := storeOraclePubKey(oraclePubKey, oracleKeyRemoteReport, conf.AbsOraclePubKeyPath()); err != nil {
				log.Errorf("failed to save oracle pub key and its remote report: %v", err)
				return err
			}

			// initialize query client using trustedBlockInfo
			queryClient, err := panacea.NewVerifiedQueryClient(context.Background(), conf, trustedBlockInfo, sgx)
			if err != nil {
				return fmt.Errorf("failed to initialize verifiedQueryClient: %w", err)
			}
			defer queryClient.Close()

			return nil
		},
	}

	// The reason why this trust block info is required is explained in the oracle.proto
	cmd.Flags().Int64(flags.FlagTrustedBlockHeight, 0, "Trusted block height")
	cmd.Flags().String(flags.FlagTrustedBlockHash, "", "Trusted block hash")
	_ = cmd.MarkFlagRequired(flags.FlagTrustedBlockHeight)
	_ = cmd.MarkFlagRequired(flags.FlagTrustedBlockHash)

	return cmd
}

// storeOraclePubKey stores base64-encoded oracle public key and its remote report
func storeOraclePubKey(oraclePubKey, oracleKeyRemoteReport []byte, filePath string) error {
	oraclePubKeyData := OraclePubKeyInfo{
		PublicKeyBase64:    base64.StdEncoding.EncodeToString(oraclePubKey),
		RemoteReportBase64: base64.StdEncoding.EncodeToString(oracleKeyRemoteReport),
	}

	oraclePubKeyFile, err := json.Marshal(oraclePubKeyData)
	if err != nil {
		return fmt.Errorf("failed to marshal oracle pub key data: %w", err)
	}

	err = os.WriteFile(filePath, oraclePubKeyFile, 0644)
	if err != nil {
		return fmt.Errorf("failed to write oracle pub key file: %w", err)
	}

	return nil
}

// getTrustedBlockInfo gets trusted block height and hash from cmd flags
func getTrustedBlockInfo(cmd *cobra.Command) (*panacea.TrustedBlockInfo, error) {
	trustedBlockHeight, err := cmd.Flags().GetInt64(flags.FlagTrustedBlockHeight)
	if err != nil {
		return nil, err
	}
	if trustedBlockHeight == 0 {
		return nil, fmt.Errorf("trusted block height cannot be zero")
	}

	trustedBlockHashStr, err := cmd.Flags().GetString(flags.FlagTrustedBlockHash)
	if err != nil {
		return nil, err
	}
	if trustedBlockHashStr == "" {
		return nil, fmt.Errorf("trusted block hash cannot be empty")
	}

	trustedBlockHash, err := hex.DecodeString(trustedBlockHashStr)
	if err != nil {
		return nil, err
	}

	return &panacea.TrustedBlockInfo{
		TrustedBlockHeight: trustedBlockHeight,
		TrustedBlockHash:   trustedBlockHash,
	}, nil
}
