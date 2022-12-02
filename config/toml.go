package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const DefaultConfigTemplate = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

###############################################################################
###                           Base Configuration                            ###
###############################################################################

log-level = "{{ .BaseConfig.LogLevel }}"
oracle-mnemonic = "{{ .BaseConfig.OracleMnemonic }}"
oracle-acc-num = "{{ .BaseConfig.OracleAccNum }}"
oracle-acc-index = "{{ .BaseConfig.OracleAccIndex }}"
listen_addr = "{{ .BaseConfig.ListenAddr }}"
data_dir = "{{ .BaseConfig.DataDir }}"

oracle_priv_key_file = "{{ .BaseConfig.OraclePrivKeyFile }}"
oracle_pub_key_file = "{{ .BaseConfig.OraclePubKeyFile }}"

###############################################################################
###                         Panacea Configuration                           ###
###############################################################################

[panacea]

chain-id = "{{ .Panacea.ChainID }}"
grpc-addr = "{{ .Panacea.GRPCAddr }}"
rpc-addr = "{{ .Panacea.RPCAddr }}"
default-gas-limit = "{{ .Panacea.DefaultGasLimit }}"
default-fee-amount = "{{ .Panacea.DefaultFeeAmount }}"

# A primary RPC address for light client verification

light-client-primary-addr = "{{ .Panacea.LightClientPrimaryAddr }}"

# Witness addresses (comma-separated) for light client verification

light-client-witness-addrs= "{{ StringsJoin .Panacea.LightClientWitnessAddrs "," }}"

# Setting log information for light client

light-client-log-level = "{{ .Panacea.LightClientLogLevel }}"

###############################################################################
###                         Ipfs Configuration                           ###
###############################################################################

[ipfs]

ipfs-node-addr = "{{ .Ipfs.IpfsNodeAddr }}"
`

var configTemplate *template.Template

// init is run automatically when the package is loaded.
func init() {
	tmpl := template.New("configFileTemplate").Funcs(template.FuncMap{
		"StringsJoin": strings.Join,
	})

	var err error
	if configTemplate, err = tmpl.Parse(DefaultConfigTemplate); err != nil {
		log.Panic(err)
	}
}

func WriteConfigTOML(path string, config *Config) error {
	var buffer bytes.Buffer
	if err := configTemplate.Execute(&buffer, config); err != nil {
		return fmt.Errorf("failed to populate config template: %w", err)
	}

	return os.WriteFile(path, buffer.Bytes(), 0600)
}

func ReadConfigTOML(path string) (*Config, error) {
	fileExt := filepath.Ext(path)

	v := viper.New()
	v.AddConfigPath(filepath.Dir(path))
	v.SetConfigName(strings.TrimSuffix(filepath.Base(path), fileExt))
	v.SetConfigType(fileExt[1:]) // excluding the dot

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var conf Config
	if err := v.Unmarshal(&conf); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := conf.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &conf, nil
}
