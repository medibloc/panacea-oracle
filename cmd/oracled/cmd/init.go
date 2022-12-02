package cmd

import (
	"fmt"
	"github.com/medibloc/panacea-oracle/client/flags"
	"github.com/medibloc/panacea-oracle/config"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

const (
	configFileName = "config.toml"
)

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configs in home dir",
		RunE: func(cmd *cobra.Command, args []string) error {
			homeDir, err := cmd.Flags().GetString(flags.FlagHome)
			if err != nil {
				return fmt.Errorf("failed to read a home flag: %w", err)
			}

			if _, err := os.Stat(homeDir); err == nil {
				return fmt.Errorf("home dir(%v) already exists", homeDir)
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("failed to check home dir: %w", err)
			}

			if err := os.MkdirAll(homeDir, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create config dir: %w", err)
			}

			defaultConfig := config.DefaultConfig()
			defaultConfig.SetHomeDir(homeDir)

			if err = os.MkdirAll(defaultConfig.AbsDataDirPath(), 0755); err != nil {
				return fmt.Errorf("failed to create db dir: %w", err)
			}

			return config.WriteConfigTOML(getConfigPath(homeDir), defaultConfig)
		},
	}
	return cmd
}

func getConfigPath(homeDir string) string {
	return filepath.Join(homeDir, configFileName)
}
