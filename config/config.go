package config

import (
	"errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"path/filepath"
)

type Config struct {
	BaseConfig `mapstructure:",squash"`

	Panacea PanaceaConfig `mapstructure:"panacea"`

	Ipfs IpfsConfig `mapstructure:"ipfs"`
}

type BaseConfig struct {
	homeDir string // not read from toml file

	LogLevel       string `mapstructure:"log-level"`
	OracleMnemonic string `mapstructure:"oracle-mnemonic"`
	OracleAccNum   uint32 `mapstructure:"oracle-acc-num"`
	OracleAccIndex uint32 `mapstructure:"oracle-acc-index"`
	ListenAddr     string `mapstructure:"listen_addr"`
	Subscriber     string `mapstructure:"subscriber"`
	DataDir        string `mapstructure:"data_dir"`

	OraclePrivKeyFile string `mapstructure:"oracle_priv_key_file"`
	OraclePubKeyFile  string `mapstructure:"oracle_pub_key_file"`
}

type PanaceaConfig struct {
	GRPCAddr                string   `mapstructure:"grpc-addr"`
	RPCAddr                 string   `mapstructure:"rpc-addr"`
	ChainID                 string   `mapstructure:"chain-id"`
	DefaultGasLimit         uint64   `mapstructure:"default-gas-limit"`
	DefaultFeeAmount        string   `mapstructure:"default-fee-amount"`
	LightClientPrimaryAddr  string   `mapstructure:"light-client-primary-addr"`
	LightClientWitnessAddrs []string `mapstructure:"light-client-witness-addrs"`
	LightClientLogLevel     string   `mapstructure:"light-client-log-level"`
}

type IpfsConfig struct {
	IpfsNodeAddr string `mapstructure:"ipfs-node-addr"`
}

func DefaultConfig() *Config {
	return &Config{
		BaseConfig: BaseConfig{
			LogLevel:       "info",
			OracleMnemonic: "",
			OracleAccNum:   0,
			OracleAccIndex: 0,
			ListenAddr:     "127.0.0.1:8080",
			DataDir:        "data",

			OraclePrivKeyFile: "oracle_priv_key.sealed",
			OraclePubKeyFile:  "oracle_pub_key.json",
		},
		Panacea: PanaceaConfig{
			GRPCAddr:                "http://127.0.0.1:9090",
			RPCAddr:                 "tcp://127.0.0.1:26657",
			ChainID:                 "",
			DefaultGasLimit:         400000,
			DefaultFeeAmount:        "2000000umed",
			LightClientPrimaryAddr:  "tcp://127.0.0.1:26657",
			LightClientWitnessAddrs: []string{"tcp://127.0.0.1:26657"},
			LightClientLogLevel:     "error",
		},
		Ipfs: IpfsConfig{
			IpfsNodeAddr: "127.0.0.1:5001",
		},
	}
}

func (c *Config) validate() error {
	if _, err := sdk.ParseCoinsNormalized(c.Panacea.DefaultFeeAmount); err != nil {
		return err
	}

	if c.Panacea.ChainID == "" {
		return errors.New("chain id should not be empty")
	}

	return nil
}

func (c *Config) SetHomeDir(dir string) {
	c.homeDir = dir
}

func (c *Config) AbsDataDirPath() string {
	return rootify(c.DataDir, c.homeDir)
}

func (c *Config) AbsOraclePrivKeyPath() string {
	return rootify(c.OraclePrivKeyFile, c.homeDir)
}

func (c *Config) AbsOraclePubKeyPath() string {
	return rootify(c.OraclePubKeyFile, c.homeDir)
}

func rootify(path, root string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}
