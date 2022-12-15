package cmd

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-oracle/client/flags"
	"github.com/medibloc/panacea-oracle/panacea"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func updateOracleInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-oracle-info",
		Short: "Update an oracle's info",
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := loadConfigFromHome(cmd)
			if err != nil {
				return err
			}

			queryClient, err := panacea.LoadVerifiedQueryClient(context.Background(), conf)
			if err != nil {
				return fmt.Errorf("failed to initialize QueryClient: %w", err)
			}
			defer queryClient.Close()

			oracleAccount, err := panacea.NewOracleAccount(conf.OracleMnemonic, conf.OracleAccNum, conf.OracleAccIndex)
			if err != nil {
				return fmt.Errorf("failed to get oracle account from mnemonic: %w", err)
			}

			oracleEndPoint, err := cmd.Flags().GetString(flags.FlagOracleEndpoint)
			if err != nil {
				return err
			}

			oracleCommissionRateStr, err := cmd.Flags().GetString(flags.FlagOracleCommissionRate)
			if err != nil {
				return err
			}

			oracleCommissionRate, err := sdk.NewDecFromStr(oracleCommissionRateStr)
			if err != nil {
				return err
			}

			//TODO: The argument of NewMsgUpdateOracleInfo will be changed when https://github.com/medibloc/panacea-core/pull/540 is merged.
			msgUpdateOracleInfo := oracletypes.NewMsgUpdateOracleInfo(oracleAccount.GetAddress(), oracleEndPoint, &oracleCommissionRate)
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

			txBytes, err := txBuilder.GenerateSignedTxBytes(oracleAccount.GetPrivKey(), conf.Panacea.DefaultGasLimit, defaultFeeAmount, msgUpdateOracleInfo)
			if err != nil {
				return fmt.Errorf("failed to generate signed Tx bytes: %w", err)
			}

			resp, err := cli.BroadcastTx(txBytes)
			if err != nil {
				return fmt.Errorf("failed to broadcast transaction: %w", err)
			}

			if resp.TxResponse.Code != 0 {
				return fmt.Errorf("update oracle info transaction failed: %v", resp.TxResponse.RawLog)
			}

			log.Infof("update-oracle-info transaction succeed. height(%v), hash(%s)", resp.TxResponse.Height, resp.TxResponse.TxHash)
			return nil
		},
	}

	cmd.Flags().String(flags.FlagOracleEndpoint, "", "endpoint of oracle")
	cmd.Flags().String(flags.FlagOracleCommissionRate, "", "oracle commission rate")

	return cmd
}
