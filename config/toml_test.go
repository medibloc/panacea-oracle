package config_test

import (
	"os"
	"testing"

	"github.com/medibloc/panacea-oracle/config"
	"github.com/stretchr/testify/require"
)

func TestWriteAndReadConfigTOML(t *testing.T) {
	path := "./config.toml"

	defaultConf := config.DefaultConfig()
	defaultConf.Panacea.ChainID = "test"
	err := config.WriteConfigTOML(path, defaultConf)
	require.NoError(t, err)
	defer os.Remove(path)

	conf, err := config.ReadConfigTOML(path)
	require.NoError(t, err)
	require.EqualValues(t, defaultConf, conf)
}
