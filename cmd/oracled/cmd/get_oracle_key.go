package cmd

import (
	"fmt"

	oracleservice "github.com/medibloc/panacea-oracle/service/oracle"
	"github.com/spf13/cobra"
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

			svc, err := oracleservice.New(conf)
			if err != nil {
				return fmt.Errorf("failed to create service: %w", err)
			}
			defer svc.Close()

			if err := svc.GetAndStoreOraclePrivKey(); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}