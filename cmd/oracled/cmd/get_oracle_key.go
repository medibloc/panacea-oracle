package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/spf13/cobra"
	tos "github.com/tendermint/tendermint/libs/os"
)

func getOracleKeyCmd() {
	cmd := &cobra.Command{
		Use:   "get-oracle-key",
		Short: "Get a shared oracle private key",
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := loadConfigFromHome(cmd)
			if err != nil {
				return err
			}

			ctx := context.Background()

			nodePrivKeyPath := conf.AbsNodePrivKeyPath()
			if !tos.FileExists(nodePrivKeyPath) {
				return errors.New("no node_priv_key.sealed file")
			}

			// get oracle account from mnemonic.
			oracleAccount, err := panacea.NewOracleAccount(conf.OracleMnemonic, conf.OracleAccNum, conf.OracleAccIndex)
			if err != nil {
				return fmt.Errorf("failed to get oracle account from mnemonic: %w", err)
			}

			// get OracleRegistration from Panacea
			queryClient, err := panacea.LoadVerifiedQueryClient(ctx, conf)
			if err != nil {
				return fmt.Errorf("failed to get queryClient: %w", err)
			}
			defer queryClient.Close()

			queryClient.GetApproveOracleRegistrationFromEvent(oracleAccount.GetAddress())

		},
	}
}
