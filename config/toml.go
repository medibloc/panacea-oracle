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

# Setting log information for light client

light-client-log-level = "{{ .Panacea.LightClientLogLevel }}"

###############################################################################
###                         IPFS Configuration                           ###
###############################################################################

[ipfs]

ipfs-node-addr = "{{ .IPFS.IPFSNodeAddr }}"

###############################################################################
###                         GRPC Configuration                           ###
###############################################################################

[grpc]
# It is address of the gRPC server. 
listen-addr = "{{ .GRPC.ListenAddr }}"

# It is a set the timeout for connection establishment for all new connections.
connection-timeout = "{{ .GRPC.ConnectionTimeout }}"

# It is the maximum number of connections to the server.
max-connections = "{{ .GRPC.MaxConnections }}"

# It will be apply a limit on the number of concurrent streams to each ServerTransport.
max-concurrent-streams = "{{ .GRPC.MaxConcurrentStreams }}"

# It is the max message size in bytes the server can receive.
max-recv-msg-size = "{{ .GRPC.MaxRecvMsgSize }}"

# It is a duration for the amount of time after which an idle connection would be closed by sending a GoAway
# Idleness duration is defined since the most recent time the number of outstanding RPCs became zero or the connection establishment.
keepalive-max-connection-idle = "{{ .GRPC.KeepaliveMaxConnectionIdle }}"

# It is a duration for the maximum amount of time a connection may exist before it will be closed by sending a GoAway.
keepalive-max-connection-age = "{{ .GRPC.KeepaliveMaxConnectionAge }}"

# It is an additive period after keepalive-max-connection-age after which the connection will be forcibly closed.
keepalive-max-connection-age-grace = "{{ .GRPC.KeepaliveMaxConnectionAgeGrace }}"

# After a duration of this time if the server doesn't see any activity it pings the client to see if the transport is still alive.
keepalive-time = "{{ .GRPC.KeepaliveTime }}"

# After having pinged for keepalive check, the server waits for a duration of Timeout and if no activity is seen even after that the connection is closed.
keepalive-timeout = "{{ .GRPC.KeepaliveTimeout }}"

# It is a set the throughput per second.
# If the throughput per second is exceeded, the client waits for 'rate-limit-wait-timeout' time and receives a failure response.
rate-limits = "{{ .GRPC.RateLimits }}"

# It is a set the waiting time when the throughput per second is exceeded (in seconds).
rate-limit-wait-timeout = "{{ .GRPC.RateLimitWaitTimeout }}"

###############################################################################
###                         API Configuration                           ###
###############################################################################

[api]

# It is only available if grpc server is enabled.
enabled = "{{ .API.Enabled }}"

# It is the address of the API server.
listen-addr = "{{ .API.ListenAddr }}"

# It is the connection timeout setting of the client used for proxy with grpc.
grpc-connect-timeout = "{{ .API.GrpcConnectTimeout }}"

# It is the maximum duration before timing out writes of the response.
write-timeout = "{{ .API.WriteTimeout }}"

# It is the maximum duration for reading the entire request, including the body.
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
