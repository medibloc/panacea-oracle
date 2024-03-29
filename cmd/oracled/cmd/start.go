package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	oracleevent "github.com/medibloc/panacea-oracle/event/oracle"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/server"
	"github.com/medibloc/panacea-oracle/service"
	"github.com/medibloc/panacea-oracle/sgx"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func startCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start oracle daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := loadConfigFromHome(cmd)
			if err != nil {
				return err
			}

			sgx := sgx.NewOracleSGX()

			queryClient, err := panacea.LoadVerifiedQueryClient(context.Background(), conf, sgx)
			if err != nil {
				return fmt.Errorf("failed to load query client: %w", err)
			}
			defer queryClient.Close()

			svc, err := service.New(conf, sgx, queryClient)
			if err != nil {
				return fmt.Errorf("failed to create service: %w", err)
			}
			defer svc.Close()

			err = svc.StartSubscriptions(
				oracleevent.NewRegisterOracleEvent(svc),
				oracleevent.NewUpgradeOracleEvent(svc),
			)
			if err != nil {
				return fmt.Errorf("failed to start event subscription: %w", err)
			}

			servers, errChan := server.Serve(svc)

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

			select {
			case err := <-errChan:
				if err != nil {
					log.Errorf("rpc server was closed with an error: %v", err)
				}
			case <-sigChan:
				log.Info("signal detected")
			}

			for _, svr := range servers {
				if err := svr.Close(); err != nil {
					log.Warnf("error occurs while server close: %v", err)
				}
			}

			return nil
		},
	}

	return cmd
}
