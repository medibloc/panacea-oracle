package cmd

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cosmos/cosmos-sdk/client/input"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/edgelesssys/ego/enclave"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-oracle/client/flags"
	"github.com/medibloc/panacea-oracle/config"
	"github.com/medibloc/panacea-oracle/crypto"
	oracleevent "github.com/medibloc/panacea-oracle/event/oracle"
	"github.com/medibloc/panacea-oracle/panacea"
	oracleservice "github.com/medibloc/panacea-oracle/service/oracle"
	"github.com/medibloc/panacea-oracle/sgx"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	tos "github.com/tendermint/tendermint/libs/os"
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

			if err := sendTxRegisterOracle(cmd, conf); err != nil {
				return fmt.Errorf("failed to send tx RegisterOracle. %w", err)
			}

			if err := subscribeApproveOracleRegistrationEvent(conf); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().Int64(flags.FlagTrustedBlockHeight, 0, "Trusted block height")
	cmd.Flags().String(flags.FlagTrustedBlockHash, "", "Trusted block hash")
	cmd.Flags().String(flags.FlagOracleEndpoint, "", "endpoint of oracle")
	cmd.Flags().String(flags.FlagOracleCommissionRate, "0.1", "oracle commission rate")
	if err := cmd.MarkFlagRequired(flags.FlagTrustedBlockHeight); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(flags.FlagTrustedBlockHash); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(flags.FlagOracleCommissionRate); err != nil {
		panic(err)
	}

	return cmd
}

func sendTxRegisterOracle(cmd *cobra.Command, conf *config.Config) error {
	// get oracle account from mnemonic.
	oracleAccount, err := panacea.NewOracleAccount(conf.OracleMnemonic, conf.OracleAccNum, conf.OracleAccIndex)
	if err != nil {
		return fmt.Errorf("failed to get oracle account from mnemonic: %w", err)
	}

	// get trusted block information
	trustedBlockInfo, err := getTrustedBlockInfo(cmd)
	if err != nil {
		return err
	}

	// initialize query client using trustedBlockInfo
	queryClient, err := panacea.NewVerifiedQueryClient(context.Background(), conf, *trustedBlockInfo)
	if err != nil {
		return fmt.Errorf("failed to initialize QueryClient: %w", err)
	}
	defer queryClient.Close()

	msgRegisterOracle, err := generateMsgRegisterOracle(cmd, conf, oracleAccount, trustedBlockInfo)
	if err != nil {
		return err
	}

	txBuilder := panacea.NewTxBuilder(queryClient)
	cli, err := panacea.NewGRPCClient(conf.Panacea.GRPCAddr)
	if err != nil {
		return fmt.Errorf("failed to generate gRPC client: %w", err)
	}
	defer cli.Close()

	defaultFeeAmount, _ := sdk.ParseCoinsNormalized(conf.Panacea.DefaultFeeAmount)
	txBytes, err := txBuilder.GenerateSignedTxBytes(
		oracleAccount.GetPrivKey(),
		conf.Panacea.DefaultGasLimit,
		defaultFeeAmount,
		msgRegisterOracle,
	)
	if err != nil {
		return fmt.Errorf("failed to generate signed Tx bytes: %w", err)
	}

	resp, err := cli.BroadcastTx(txBytes)
	if err != nil {
		return fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	if resp.TxResponse.Code != 0 {
		return fmt.Errorf("register oracle transaction failed: %v", resp.TxResponse.RawLog)
	}

	log.Infof("register-oracle transaction succeed. height(%v), hash(%s)", resp.TxResponse.Height, resp.TxResponse.TxHash)

	return nil
}

func generateMsgRegisterOracle(cmd *cobra.Command, conf *config.Config, oracleAccount *panacea.OracleAccount, trustedBlockInfo *panacea.TrustedBlockInfo) (*oracletypes.MsgRegisterOracle, error) {
	// if node key exists, return error.
	nodePrivKeyPath := conf.AbsNodePrivKeyPath()
	if tos.FileExists(nodePrivKeyPath) {
		buf := bufio.NewReader(os.Stdin)
		ok, err := input.GetConfirmation("There is an existing node key. \nAre you sure to delete and re-generate node key?", buf, os.Stderr)
		if err != nil || !ok {
			log.Infof("Node key generation is canceled.")
			return nil, err
		}
	}

	// generate node key and its remote report
	nodePubKey, nodePubKeyRemoteReport, err := generateAndSealedNodeKey(nodePrivKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to generate node key pair: %w", err)
	}

	report, err := enclave.VerifyRemoteReport(nodePubKeyRemoteReport)
	if err != nil {
		return nil, fmt.Errorf("failed to verification remoteReport. %w", err)
	}
	uniqueID := hex.EncodeToString(report.UniqueID)

	oracleEndpoint, err := cmd.Flags().GetString(flags.FlagOracleEndpoint)
	if err != nil {
		return nil, err
	}

	// request register oracle Tx to Panacea
	oracleCommissionRateStr, err := cmd.Flags().GetString(flags.FlagOracleCommissionRate)
	if err != nil {
		return nil, err
	}

	oracleCommissionRate, err := sdk.NewDecFromStr(oracleCommissionRateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse oracleCommissionRate. input(%s). %w", oracleCommissionRateStr, err)
	}

	msgRegisterOracle := oracletypes.NewMsgRegisterOracle(
		uniqueID,
		oracleAccount.GetAddress(),
		nodePubKey,
		nodePubKeyRemoteReport,
		trustedBlockInfo.TrustedBlockHeight,
		trustedBlockInfo.TrustedBlockHash,
		oracleEndpoint,
		oracleCommissionRate,
	)
	return msgRegisterOracle, nil
}

// generateAndSealedNodeKey generates random node key and its remote report
// And the generated private key is sealed and stored
func generateAndSealedNodeKey(nodePrivKeyPath string) ([]byte, []byte, error) {
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

func subscribeApproveOracleRegistrationEvent(conf *config.Config) error {
	svc, err := oracleservice.New(conf)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer svc.Close()

	doneChan := make(chan error, 1)
	sigChan := make(chan os.Signal, 1)

	err = svc.StartSubscriptions(
		oracleevent.NewApproveOracleRegistrationEvent(svc, doneChan),
	)
	if err != nil {
		return fmt.Errorf("failed to start event subscription: %w", err)
	}

	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	select {
	case err := <-doneChan:
		if err != nil {
			log.Errorf("oraclePrivateKey could not be retrieved. %v", err)
		} else {
			log.Infof("oraclePrivateKey is retrieved successfully")
		}
	case <-sigChan:
		log.Info("signal detected")
	}
	return nil
}
