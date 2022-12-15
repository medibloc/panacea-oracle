package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/medibloc/panacea-oracle/sgx"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func verifyReportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify-report [report-file-path]",
		Short: "Verify whether the report was properly generated in the SGX environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// read oracle remote targetReport
			pubKeyInfo, err := readOracleRemoteReport(args[0])
			if err != nil {
				log.Errorf("failed to read remote targetReport: %v", err)
				return err
			}

			if err := verifyPubKeyRemoteReport(*pubKeyInfo); err != nil {
				log.Errorf("failed to verify the public key and its remote report: %v", err)
				return err
			}

			log.Infof("remote report is verified successfully")

			return nil
		},
	}

	return cmd
}

func readOracleRemoteReport(filename string) (*OraclePubKeyInfo, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var pubKeyInfo OraclePubKeyInfo

	if err := json.Unmarshal(file, &pubKeyInfo); err != nil {
		return nil, err
	}

	return &pubKeyInfo, nil
}

func verifyPubKeyRemoteReport(pubKeyInfo OraclePubKeyInfo) error {
	pubKey, err := base64.StdEncoding.DecodeString(pubKeyInfo.PublicKeyBase64)
	if err != nil {
		return fmt.Errorf("failed to decode oracle public key: %w", err)
	}

	targetReport, err := base64.StdEncoding.DecodeString(pubKeyInfo.RemoteReportBase64)
	if err != nil {
		return fmt.Errorf("failed to decode oracle public key remote report: %w", err)
	}

	selfEnclaveInfo, err := sgx.GetSelfEnclaveInfo()
	if err != nil {
		return fmt.Errorf("failed to set self-enclave info: %w", err)
	}

	// verify remote report
	if err := sgx.VerifyRemoteReport(targetReport, pubKey, *selfEnclaveInfo); err != nil {
		return fmt.Errorf("failed to verify report: %w", err)
	}

	return nil
}
