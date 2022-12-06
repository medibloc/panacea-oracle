package cmd

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/edgelesssys/ego/enclave"
	"github.com/medibloc/panacea-oracle/client/flags"
	"github.com/medibloc/panacea-oracle/crypto"
	oracleevent "github.com/medibloc/panacea-oracle/event/oracle"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/service"
	"github.com/medibloc/panacea-oracle/sgx"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	tos "github.com/tendermint/tendermint/libs/os"
	"os"
)

const (
	flagOracleEndpoint       = "oracle-endpoint"
	flagOracleDescription    = "oracle-description"
	flagOracleCommissionRate = "oracle-commission-rate"
)

func registerOracleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-oracle",
		Short: "Register an oracle",
		RunE: func(cmd *cobra.Command, args []string) error {
			// load config
			conf, err := loadConfigFromHome(cmd)
			if err != nil {
				return err
			}

			// if node key exists, return error.
			nodePrivKeyPath := conf.AbsNodePrivKeyPath()
			if tos.FileExists(nodePrivKeyPath) {
				buf := bufio.NewReader(os.Stdin)
				ok, err := input.GetConfirmation("There is an existing node key. \nAre you sure to delete and re-generate node key?", buf, os.Stderr)
				if err != nil || !ok {
					log.Printf("Node key generation is canceled.")
					return err
				}
			}

			// get trusted block information
			trustedBlockInfo, err := getTrustedBlockInfo(cmd)
			if err != nil {
				return fmt.Errorf("failed to get trusted block info: %w", err)
			}

			// initialize query client using trustedBlockInfo
			queryClient, err := panacea.NewQueryClient(context.Background(), conf, *trustedBlockInfo)
			if err != nil {
				return fmt.Errorf("failed to initialize QueryClient: %w", err)
			}
			defer queryClient.Close()

			// get oracle account from mnemonic.
			_, err = panacea.NewOracleAccount(conf.OracleMnemonic, conf.OracleAccNum, conf.OracleAccIndex)
			if err != nil {
				return fmt.Errorf("failed to get oracle account from mnemonic: %w", err)
			}

			// generate node key and its remote report
			_, nodePubKeyRemoteReport, err := generateSealedNodeKey(nodePrivKeyPath)
			if err != nil {
				return fmt.Errorf("failed to generate node key pair: %w", err)
			}

			report, _ := enclave.VerifyRemoteReport(nodePubKeyRemoteReport)
			_ = hex.EncodeToString(report.UniqueID)

			// request register oracle Tx to Panacea
			// TODO: add MsgRegisterOracle Tx

			// subscribe approve oracle registration event
			svc, err := service.New(conf)
			if err != nil {
				return fmt.Errorf("failed to create service: %w", err)
			}
			defer svc.Close()

			err = svc.StartSubscriptions(
				oracleevent.NewApproveOracleRegistrationEvent(svc),
			)
			if err != nil {
				return fmt.Errorf("failed to start event subscription: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().Int64(flags.FlagTrustedBlockHeight, 0, "Trusted block height")
	cmd.Flags().String(flags.FlagTrustedBlockHash, "", "Trusted block hash")
	cmd.Flags().String(flagOracleEndpoint, "", "endpoint of oracle")
	cmd.Flags().String(flagOracleDescription, "", "description of oracle")
	cmd.Flags().String(flagOracleCommissionRate, "0.1", "oracle commission rate")
	_ = cmd.MarkFlagRequired(flags.FlagTrustedBlockHeight)
	_ = cmd.MarkFlagRequired(flags.FlagTrustedBlockHash)

	return cmd
}

// generateSealedNodeKey generates random node key and its remote report
// And the generated private key is sealed and stored
func generateSealedNodeKey(nodePrivKeyPath string) ([]byte, []byte, error) {
	nodePrivKey, err := crypto.NewPrivKey()
	if err != nil {
		return nil, nil, err
	}

	if err := sgx.SealToFile(nodePrivKey.Serialize(), nodePrivKeyPath); err != nil {
		return nil, nil, err
	}

	nodePubKey := nodePrivKey.PubKey().SerializeCompressed()
	oraclePubKeyHash := sha256.Sum256(nodePubKey)
	nodeKeyRemoteReport, err := sgx.GenerateRemoteReport(oraclePubKeyHash[:])
	if err != nil {
		return nil, nil, err
	}

	return nodePubKey, nodeKeyRemoteReport, nil
}
