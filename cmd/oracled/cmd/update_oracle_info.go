package cmd

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-oracle/client/flags"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/service"
	"github.com/medibloc/panacea-oracle/sgx"
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

			sgx := sgx.NewOracleSgx()

			queryClient, err := panacea.LoadVerifiedQueryClient(context.Background(), conf, sgx)
			if err != nil {
				return fmt.Errorf("failed to load query client: %w", err)
			}

			svc, err := service.New(conf, sgx, queryClient)
			if err != nil {
				return fmt.Errorf("failed to create service: %w", err)
			}
			defer svc.Close()

			oracleAccount := svc.GetOracleAcc()
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

			msgUpdateOracleInfo := oracletypes.NewMsgUpdateOracleInfo(oracleAccount.GetAddress(), oracleEndPoint, &oracleCommissionRate)
			txHeight, txHash, err := svc.BroadcastTx(msgUpdateOracleInfo)
			if err != nil {
				return err
			}

			log.Infof("update-oracle-info transaction succeed. height(%v), hash(%s)", txHeight, txHash)
			return nil
		},
	}

	cmd.Flags().String(flags.FlagOracleEndpoint, "", "endpoint of oracle")
	cmd.Flags().String(flags.FlagOracleCommissionRate, "", "oracle commission rate")

	return cmd
}
