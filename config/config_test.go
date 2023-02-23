package config_test

import (
	"path/filepath"
	"testing"

	"github.com/medibloc/panacea-oracle/config"
	"github.com/stretchr/testify/require"
)

// TestConfigAbsPath performs a test on the Abs function.
func TestConfigAbsPath(t *testing.T) {
	conf := config.DefaultConfig()

	defaultHomeDir := "/home/oracle"

	conf.SetHomeDir(defaultHomeDir)
	require.Equal(t, filepath.Join(defaultHomeDir, conf.OraclePrivKeyFile), conf.AbsOraclePrivKeyPath())
	require.Equal(t, filepath.Join(defaultHomeDir, conf.OraclePubKeyFile), conf.AbsOraclePubKeyPath())
	require.Equal(t, filepath.Join(defaultHomeDir, conf.DataDir), conf.AbsDataDirPath())
	require.Equal(t, filepath.Join(defaultHomeDir, conf.NodePrivKeyFile), conf.AbsNodePrivKeyPath())

	newHomeDir := "/home/new_oracle"
	conf.OraclePrivKeyFile = filepath.Join(newHomeDir, "oracle_priv_key.sealed")
	require.Equal(t, filepath.Join(newHomeDir, "oracle_priv_key.sealed"), conf.AbsOraclePrivKeyPath())
	conf.OraclePubKeyFile = filepath.Join(newHomeDir, "oracle_pub_key.json")
	require.Equal(t, filepath.Join(newHomeDir, "oracle_pub_key.json"), conf.AbsOraclePubKeyPath())
	conf.DataDir = filepath.Join(newHomeDir, "data")
	require.Equal(t, filepath.Join(newHomeDir, "data"), conf.AbsDataDirPath())
	conf.NodePrivKeyFile = filepath.Join(newHomeDir, "node_priv_key.sealed")
	require.Equal(t, filepath.Join(newHomeDir, "node_priv_key.sealed"), conf.AbsNodePrivKeyPath())
}
