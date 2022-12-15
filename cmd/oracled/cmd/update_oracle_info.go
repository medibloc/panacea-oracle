package cmd

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-oracle/client/flags"
	"github.com/medibloc/panacea-oracle/service"
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

			svc, err := service.New(conf)
			if err != nil {
				return err
			}
			defer svc.Close()

			oracleAccount := svc.OracleAcc()
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
