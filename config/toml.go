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
data-dir = "{{ .BaseConfig.DataDir }}"

oracle-priv-key-file = "{{ .BaseConfig.OraclePrivKeyFile }}"
oracle-pub-key-file = "{{ .BaseConfig.OraclePubKeyFile }}"
node-priv-key-file = "{{ .BaseConfig.NodePrivKeyFile }}"

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

# Log level for light client

light-client-log-level = "{{ .Panacea.LightClientLogLevel }}"

###############################################################################
###                         GRPC Configuration                           ###
###############################################################################

[grpc]
# TCP or UNIX socket address for the gRPC server to listen on
listen-addr = "{{ .GRPC.ListenAddr }}"

# Timeout for connection establishment for all new connections.
connection-timeout = "{{ .GRPC.ConnectionTimeout }}"

# Maximum number of simultaneous connections that the gRPC server can handle
max-connections = "{{ .GRPC.MaxConnections }}"

# Maximum number of concurrent streams that each gRPC connection can handle
max-concurrent-streams = "{{ .GRPC.MaxConcurrentStreams }}"

# Max message size in bytes the server can receive.
max-recv-msg-size = "{{ .GRPC.MaxRecvMsgSize }}"

# Duration for the amount of time after which an idle connection would be closed by sending a GoAway.
# Idleness duration is defined since the most recent time the number of outstanding RPCs became zero or the connection establishment.
keepalive-max-connection-idle = "{{ .GRPC.KeepaliveMaxConnectionIdle }}"

# Duration for the maximum amount of time a connection may exist before it will be closed by sending a GoAway.
keepalive-max-connection-age = "{{ .GRPC.KeepaliveMaxConnectionAge }}"

# Additive period after keepalive-max-connection-age after which the connection will be forcibly closed.
keepalive-max-connection-age-grace = "{{ .GRPC.KeepaliveMaxConnectionAgeGrace }}"

# After a duration of this time if the server doesn't see any activity it pings the client to see if the transport is still alive.
keepalive-time = "{{ .GRPC.KeepaliveTime }}"

# After having pinged for keepalive check, the server waits for a duration of Timeout and if no activity is seen even after that the connection is closed.
keepalive-timeout = "{{ .GRPC.KeepaliveTimeout }}"

# Max throughput per second that the server can handle.
# If the throughput per second is exceeded, the request is blocked for up to 'rate-limit-wait-timeout'. If the request is not processed after that, an error is returned to the client.
rate-limits = "{{ .GRPC.RateLimits }}"

# Timeout to wait if a request is blocked due to rate limiting.
rate-limit-wait-timeout = "{{ .GRPC.RateLimitWaitTimeout }}"

###############################################################################
###                         API Configuration                           ###
###############################################################################

[api]

# Enable the REST API server (gRPC gateway)
enabled = "{{ .API.Enabled }}"

# TCP or UNIX socket address for the API server to listen on
listen-addr = "{{ .API.ListenAddr }}"

# Timeout for the API server to establish connections with the underlying gRPC server
grpc-connect-timeout = "{{ .API.GrpcConnectTimeout }}"

# Maximum duration before timing out writes of the response.
write-timeout = "{{ .API.WriteTimeout }}"

# Maximum duration for reading the entire request, including the body.
read-timeout = "{{ .API.ReadTimeout }}"
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
