package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/medibloc/panacea-oracle/server/rpc"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func startCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start oracle daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfigFromHome(cmd)
			if err != nil {
				return err
			}

			/*svc, err := service.New(conf)
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
			}*/

			errChan := make(chan error, 1)
			sigChan := make(chan os.Signal, 1)

			/*srv := server.New(svc)

			go func() {
				if err := srv.Run(); err != nil {
					if !errors.Is(err, http.ErrServerClosed) {
						errChan <- err
					} else {
						close(errChan)
					}
				}
			}()
			*/

			if err := rpc.Serve(cfg, errChan); err != nil {
				errChan <- err
			}

			signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

			select {
			case err := <-errChan:
				if err != nil {
					log.Errorf("http server was closed with an error: %v", err)
				}
			case <-sigChan:
				log.Info("signal detected")
			}

			log.Infof("starting graceful shutdown")

			/*ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			if err := srv.Shutdown(ctxTimeout); err != nil {
				return fmt.Errorf("error occurs while server shutting down: %w", err)
			}*/

			return nil
		},
	}

	return cmd
}
