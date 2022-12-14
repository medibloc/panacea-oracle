package cmd

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/client/input"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/edgelesssys/ego/enclave"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-oracle/client/flags"
	"github.com/medibloc/panacea-oracle/crypto"
	oracleevent "github.com/medibloc/panacea-oracle/event/oracle"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/sgx"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	tos "github.com/tendermint/tendermint/libs/os"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
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
			queryClient, err := panacea.NewVerifiedQueryClient(context.Background(), conf, *trustedBlockInfo)
			if err != nil {
				return fmt.Errorf("failed to initialize QueryClient: %w", err)
			}
			defer queryClient.Close()

			// get oracle account from mnemonic.
			oracleAccount, err := panacea.NewOracleAccount(conf.OracleMnemonic, conf.OracleAccNum, conf.OracleAccIndex)
			if err != nil {
				return fmt.Errorf("failed to get oracle account from mnemonic: %w", err)
			}

			// generate node key and its remote report
			nodePubKey, nodePubKeyRemoteReport, err := generateSealedNodeKey(nodePrivKeyPath)
			if err != nil {
				return fmt.Errorf("failed to generate node key pair: %w", err)
			}

			report, _ := enclave.VerifyRemoteReport(nodePubKeyRemoteReport)
			uniqueID := hex.EncodeToString(report.UniqueID)

			// request register oracle Tx to Panacea
			oracleCommissionRateStr, err := cmd.Flags().GetString(flagOracleCommissionRate)
			if err != nil {
				return err
			}

			oracleCommissionRate, err := sdk.NewDecFromStr(oracleCommissionRateStr)
			if err != nil {
				return err
			}

			endPoint, err := cmd.Flags().GetString(flagOracleEndpoint)
			if err != nil {
				return err
			}

			// TODO: OracleCommissionMaxRate & OracleCommissionMaxChangeRate will be added in other PR.
			msgRegisterOracle := oracletypes.NewMsgRegisterOracle(uniqueID, oracleAccount.GetAddress(), nodePubKey, nodePubKeyRemoteReport, trustedBlockInfo.TrustedBlockHeight, trustedBlockInfo.TrustedBlockHash, endPoint, oracleCommissionRate, sdk.NewDec(0), sdk.NewDec(0))
			txBuilder := panacea.NewTxBuilder(queryClient)
			cli, err := panacea.NewGRPCClient(conf.Panacea.GRPCAddr)
			if err != nil {
				return fmt.Errorf("failed to generate gRPC client: %w", err)
			}
			defer cli.Close()

			defaultFeeAmount, err := sdk.ParseCoinsNormalized(conf.Panacea.DefaultFeeAmount)
			if err != nil {
				return err
			}

			txBytes, err := txBuilder.GenerateSignedTxBytes(oracleAccount.GetPrivKey(), conf.Panacea.DefaultGasLimit, defaultFeeAmount, msgRegisterOracle)
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

			// subscribe approval of oracle registration and handle it
			client, err := rpchttp.New(conf.Panacea.RPCAddr, "/websocket")
			if err != nil {
				return err
			}

			if err := client.Start(); err != nil {
				return err
			}
			defer func() {
				if err = client.Stop(); err != nil {
					log.Errorf("error occurs when rpc client stops: %s", err.Error())
				}
			}()

			event := oracleevent.NewApproveOracleRegistrationEvent()

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			txs, err := client.Subscribe(ctx, "", event.GetEventQuery())
			if err != nil {
				return err
			}

			errChan := make(chan error, 1)

			for tx := range txs {
				errChan <- event.EventHandler(tx)
			}

			err = <-errChan
			if err != nil {
				log.Infof("Error occurs while getting shared oracle private key. Please retrieve it via get-oracle-key cmd: %v", err)
				return err
			} else {
				log.Infof("oracle private key is successfully shared. You can start oracle now!")
				return nil
			}
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
