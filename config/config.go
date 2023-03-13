package config

import (
	"errors"
	"path/filepath"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Config struct {
	BaseConfig `mapstructure:",squash"`

	Panacea PanaceaConfig `mapstructure:"panacea"`

	IPFS IPFSConfig `mapstructure:"ipfs"`

	API APIConfig `mapstructure:"api"`
}

type BaseConfig struct {
	homeDir string // not read from toml file

	LogLevel       string `mapstructure:"log-level"`
	OracleMnemonic string `mapstructure:"oracle-mnemonic"`
	OracleAccNum   uint32 `mapstructure:"oracle-acc-num"`
	OracleAccIndex uint32 `mapstructure:"oracle-acc-index"`
	Subscriber     string `mapstructure:"subscriber"`
	DataDir        string `mapstructure:"data-dir"`

	OraclePrivKeyFile string `mapstructure:"oracle-priv-key-file"`
	OraclePubKeyFile  string `mapstructure:"oracle-pub-key-file"`
	NodePrivKeyFile   string `mapstructure:"node-priv-key-file"`
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

type IPFSConfig struct {
	IPFSNodeAddr string `mapstructure:"ipfs-node-addr"`
}

type APIConfig struct {
	ListenAddr         string `mapstructure:"listen-addr"`
	WriteTimeout       int64  `mapstructure:"write-timeout"`
	ReadTimeout        int64  `mapstructure:"read-timeout"`
	MaxConnections     int    `mapstructure:"max-connections"`
	MaxRequestBodySize int64  `mapstructure:"max-request-body-size"`
}

func DefaultConfig() *Config {
	return &Config{
		BaseConfig: BaseConfig{
			LogLevel:       "info",
			OracleMnemonic: "",
			OracleAccNum:   0,
			OracleAccIndex: 0,
			DataDir:        "data",

			OraclePrivKeyFile: "oracle_priv_key.sealed",
			OraclePubKeyFile:  "oracle_pub_key.json",
			NodePrivKeyFile:   "node_priv_key.sealed",
		},
		Panacea: PanaceaConfig{
			GRPCAddr: "http://127.0.0.1:9090",
			RPCAddr:  "tcp://127.0.0.1:26657",
			ChainID:  "",
			// TODO: calculate fee instead of using default fee
			DefaultGasLimit:         400000,
			DefaultFeeAmount:        "2000000umed",
			LightClientPrimaryAddr:  "tcp://127.0.0.1:26657",
			LightClientWitnessAddrs: []string{"tcp://127.0.0.1:26657"},
			LightClientLogLevel:     "error",
		},
		IPFS: IPFSConfig{
			IPFSNodeAddr: "127.0.0.1:5001",
		},
		API: APIConfig{
			ListenAddr:         "127.0.0.1:8080",
			WriteTimeout:       60,
			ReadTimeout:        15,
			MaxConnections:     50,
			MaxRequestBodySize: 4 << (10 * 2), // 4MB
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

func (c *Config) AbsNodePrivKeyPath() string {
	return rootify(c.NodePrivKeyFile, c.homeDir)
}

func rootify(path, root string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}
