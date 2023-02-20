package cmd

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/edgelesssys/ego/enclave"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-oracle/client/flags"
	oracleevent "github.com/medibloc/panacea-oracle/event/oracle"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/service"
	"github.com/medibloc/panacea-oracle/sgx"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	tos "github.com/tendermint/tendermint/libs/os"
)

func upgradeOracle() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade-oracle",
		Short: "Upgrade the oracle",
		RunE: func(cmd *cobra.Command, args []string) error {
			// load config
			conf, err := loadConfigFromHome(cmd)
			if err != nil {
				return err
			}

			// get trusted block information
			trustedBlockInfo, err := getTrustedBlockInfo(cmd)
			if err != nil {
				return err
			}

			sgx := sgx.NewOracleSGX()

			queryClient, err := panacea.NewVerifiedQueryClient(context.Background(), conf, trustedBlockInfo, sgx)
			if err != nil {
				return fmt.Errorf("failed to create queryClient: %w", err)
			}

			svc, err := service.New(conf, sgx, queryClient)
			if err != nil {
				return fmt.Errorf("failed to create service: %w", err)
			}
			defer svc.Close()

			if err := sendTxUpgradeOracle(svc, trustedBlockInfo); err != nil {
				return fmt.Errorf("failed to send tx UpgradeOracle: %w", err)
			}

			if err := subscribeApproveOracleUpgradeEvent(svc); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().Int64(flags.FlagTrustedBlockHeight, 0, "Trusted block height")
	cmd.Flags().String(flags.FlagTrustedBlockHash, "", "Trusted block hash")
	if err := cmd.MarkFlagRequired(flags.FlagTrustedBlockHeight); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(flags.FlagTrustedBlockHash); err != nil {
		panic(err)
	}

	return cmd
}

func sendTxUpgradeOracle(svc service.Service, trustedBlockInfo *panacea.TrustedBlockInfo) error {
	oracleAccount := svc.OracleAcc()

	msgRegisterOracle, err := generateMsgUpgradeOracle(svc, oracleAccount, trustedBlockInfo)
	if err != nil {
		return fmt.Errorf("failed to generate MsgUpgradeOracle: %w", err)
	}

	txHeight, txHash, err := svc.BroadcastTx(msgRegisterOracle)
	if err != nil {
		return fmt.Errorf("failed to broadcast UpgradeOracle Tx: %w", err)
	}

	log.Infof("UpgradeOracle transaction succeed. height(%d), hash(%s)", txHeight, txHash)

	return nil
}

func generateMsgUpgradeOracle(svc service.Service, oracleAccount *panacea.OracleAccount, trustedBlockInfo *panacea.TrustedBlockInfo) (*oracletypes.MsgUpgradeOracle, error) {
	conf := svc.Config()

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
	nodePubKey, nodePubKeyRemoteReport, err := generateAndSealedNodeKey(svc.SGX(), nodePrivKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to generate node key pair: %w", err)
	}

	report, _ := enclave.VerifyRemoteReport(nodePubKeyRemoteReport)
	uniqueID := hex.EncodeToString(report.UniqueID)

	msgRegisterOracle := &oracletypes.MsgUpgradeOracle{
		UniqueId:               uniqueID,
		OracleAddress:          oracleAccount.GetAddress(),
		NodePubKey:             nodePubKey,
		NodePubKeyRemoteReport: nodePubKeyRemoteReport,
		TrustedBlockHeight:     trustedBlockInfo.TrustedBlockHeight,
		TrustedBlockHash:       trustedBlockInfo.TrustedBlockHash,
	}

	return msgRegisterOracle, nil
}

func subscribeApproveOracleUpgradeEvent(svc service.Service) error {
	doneChan := make(chan error, 1)
	sigChan := make(chan os.Signal, 1)

	err := svc.StartSubscriptions(
		oracleevent.NewApproveOracleUpgradeEvent(svc, doneChan),
	)
	if err != nil {
		return fmt.Errorf("failed to start event subscription: %w", err)
	}

	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	select {
	case err := <-doneChan:
		if err != nil {
			log.Errorf("failed to retrieve oracle private key: %s", err.Error())
		} else {
			log.Infof("oracle private key is retrieved successfully")
		}
	case <-sigChan:
		log.Infof("signal detected")
	}

	return nil
}
