package config

import (
	"errors"
	"path/filepath"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Config struct {
	BaseConfig `mapstructure:",squash"`

	Panacea PanaceaConfig `mapstructure:"panacea"`

	GRPC GRPCConfig `mapstructure:"grpc"`

	API APIConfig `mapstructure:"api"`

	Consumer ConsumerConfig `mapstructure:"consumer"`
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

type APIConfig struct {
	Enabled            bool          `mapstructure:"enabled"`
	ListenAddr         string        `mapstructure:"listen-addr"`
	GrpcConnectTimeout time.Duration `mapstructure:"grpc-connect-timeout"`
	WriteTimeout       time.Duration `mapstructure:"write-timeout"`
	ReadTimeout        time.Duration `mapstructure:"read-timeout"`
	MaxConnections     int           `mapstructure:"max-connections"`
	MaxRequestBodySize int64         `mapstructure:"max-request-body-size"`
}

type GRPCConfig struct {
	ListenAddr                     string        `mapstructure:"listen-addr"`
	ConnectionTimeout              time.Duration `mapstructure:"connection-timeout"`
	MaxConnections                 int           `mapstructure:"max-connections"`
	MaxConcurrentStreams           int           `mapstructure:"max-concurrent-streams"`
	MaxRecvMsgSize                 int           `mapstructure:"max-recv-msg-size"`
	KeepaliveMaxConnectionIdle     time.Duration `mapstructure:"keepalive-max-connection-idle"`
	KeepaliveMaxConnectionAge      time.Duration `mapstructure:"keepalive-max-connection-age"`
	KeepaliveMaxConnectionAgeGrace time.Duration `mapstructure:"keepalive-max-connection-age-grace"`
	KeepaliveTime                  time.Duration `mapstructure:"keepalive-time"`
	KeepaliveTimeout               time.Duration `mapstructure:"keepalive-timeout"`
	RateLimits                     int           `mapstructure:"rate-limits"`
	RateLimitWaitTimeout           time.Duration `mapstructure:"rate-limit-wait-timeout"`
}

type ConsumerConfig struct {
	Timeout time.Duration `mapstructure:"timeout"`
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
		GRPC: GRPCConfig{
			ListenAddr:                     "127.0.0.1:9090",
			ConnectionTimeout:              time.Minute * 2,
			MaxConnections:                 50,
			MaxConcurrentStreams:           0,
			MaxRecvMsgSize:                 1 << (10 * 2), // 1MB
			KeepaliveMaxConnectionIdle:     0,
			KeepaliveMaxConnectionAge:      0,
			KeepaliveMaxConnectionAgeGrace: 0,
			KeepaliveTime:                  time.Hour * 2,
			KeepaliveTimeout:               time.Second * 20,
			RateLimits:                     100,
			RateLimitWaitTimeout:           time.Second * 5,
		},
		API: APIConfig{
			Enabled:            true,
			ListenAddr:         "127.0.0.1:8080",
			GrpcConnectTimeout: time.Second * 10,
			WriteTimeout:       time.Second * 60,
			ReadTimeout:        time.Second * 15,
			MaxConnections:     50,
			MaxRequestBodySize: 1 << (10 * 2), // 1MB
		},
		Consumer: ConsumerConfig{
			Timeout: time.Second * 5,
		},
		Consumer: ConsumerConfig{
			Timeout: time.Second * 5,
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
