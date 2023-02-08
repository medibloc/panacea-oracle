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

	GRPC GRPCConfig `mapstructure:"grpc"`

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
	Enabled            bool   `mapstructure:"enabled"`
	ListenAddr         string `mapstructure:"listen-addr"`
	GrpcConnectTimeout int64  `mapstructure:"grpc-connect-timeout"`
	WriteTimeout       int64  `mapstructure:"write-timeout"`
	ReadTimeout        int64  `mapstructure:"read-timeout"`
}

type GRPCConfig struct {
	ListenAddr           string `mapstructure:"listen-addr"`
	ConnectionTimeout    int64  `mapstructure:"connection-timeout"`
	MaxConnectionSize    int    `mapstructure:"max-connection-size"`
	RateLimitPerSecond   int    `mapstructure:"rate-limit-per-second"`
	RateLimitWaitTimeout int64  `mapstructure:"rate-limit-wait-timeout"`
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
			GRPCAddr: "tcp://127.0.0.1:9090",
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
		GRPC: GRPCConfig{
			ListenAddr:           "tcp://127.0.0.1:9090",
			ConnectionTimeout:    120,
			MaxConnectionSize:    50,
			RateLimitPerSecond:   100,
			RateLimitWaitTimeout: 5,
		},
		API: APIConfig{
			Enabled:            true,
			ListenAddr:         "http://127.0.0.1:8080",
			GrpcConnectTimeout: 10,
			WriteTimeout:       60,
			ReadTimeout:        15,
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
