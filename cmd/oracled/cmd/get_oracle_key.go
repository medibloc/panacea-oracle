package cmd

import (
	"context"
	"fmt"

	"github.com/medibloc/panacea-oracle/client/flags"
	"github.com/medibloc/panacea-oracle/key"
	"github.com/medibloc/panacea-oracle/service"
	"github.com/spf13/cobra"
)

const (
	fromRegistration = "registration"
	fromUpgrade      = "upgrade"
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

			svc, err := service.New(conf)
			if err != nil {
				return err
			}

			ctx := context.Background()

			uniqueID := svc.EnclaveInfo().UniqueIDHex()
			oracleAddress := svc.OracleAcc().GetAddress()

			from, err := cmd.Flags().GetString(flags.FlagFromOracleRegistrationOrUpgrade)
			if err != nil {
				return err
			}

			switch from {
			case fromRegistration:
				oracleRegistration, err := svc.QueryClient().GetOracleRegistration(ctx, uniqueID, oracleAddress)
				if err != nil {
					return fmt.Errorf("failed to get oracle registration: %w", err)
				}

				if len(oracleRegistration.EncryptedOraclePrivKey) == 0 {
					return fmt.Errorf("the encrypted oracle private key has not set yet. please try again later")
				}
				return key.RetrieveAndStoreOraclePrivKey(ctx, svc, oracleRegistration.EncryptedOraclePrivKey)

			case fromUpgrade:
				oracleUpgrade, err := svc.QueryClient().GetOracleUpgrade(ctx, uniqueID, oracleAddress)
				if err != nil {
					return fmt.Errorf("failed to get oracle upgrade: %w", err)
				}
				if len(oracleUpgrade.EncryptedOraclePrivKey) == 0 {
					return fmt.Errorf("the encrypted oracle private key has not set yet. please try again later")
				}
				return key.RetrieveAndStoreOraclePrivKey(ctx, svc, oracleUpgrade.EncryptedOraclePrivKey)

			default:
				return fmt.Errorf("invalid --from flag input. please put \"registration\" or \"upgrade\"")
			}
		},
	}

	cmd.Flags().String(flags.FlagFromOracleRegistrationOrUpgrade, fromUpgrade, "where to get the key from")

	return cmd
}
