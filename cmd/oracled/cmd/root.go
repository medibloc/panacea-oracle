package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/medibloc/panacea-oracle/client/flags"
	"github.com/medibloc/panacea-oracle/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "oracled",
		Short: "oracle daemon",
	}
)

func Execute() error {
	return rootCmd.Execute()
}

// init is run automatically when the package is loaded.
func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	defaultAppHomeDir := filepath.Join(userHomeDir, ".oracle")

	rootCmd.PersistentFlags().String(flags.FlagHome, defaultAppHomeDir, "application home directory")

	rootCmd.AddCommand(
		initCmd(),
		genOracleKeyCmd(),
		registerOracleCmd(),
		getOracleKeyCmd(),
		updateOracleInfoCmd(),
		startCmd(),
	)
}

func initLogger(conf *config.Config) error {
	logLevel, err := log.ParseLevel(conf.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to parse log level: %w", err)
	}

	log.SetLevel(logLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})

	return nil
}

func loadConfigFromHome(cmd *cobra.Command) (*config.Config, error) {
	homeDir, err := cmd.Flags().GetString(flags.FlagHome)
	if err != nil {
		return nil, fmt.Errorf("failed to read a home flag: %w", err)
	}

	conf, err := config.ReadConfigTOML(getConfigPath(homeDir))
	if err != nil {
		return nil, fmt.Errorf("failed to read config from file: %w", err)
	}
	conf.SetHomeDir(homeDir)

	if err := initLogger(conf); err != nil {
		return nil, fmt.Errorf("failed to init logger: %w", err)
	}

	return conf, nil
}
