package cmd

import (
	"github.com/medibloc/panacea-oracle/client/flags"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
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
	)
}
