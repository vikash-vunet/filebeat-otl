package filebeatotl

import (
	"github.com/elastic/beats/v7/libbeat/outputs/codec"
	"github.com/elastic/elastic-agent-libs/config"
)

// config object structure to store all OTLP Configs
type OtlConfig struct {
	ServiceName    string           `config:"service_name"`
	ServiceVersion string           `config:"service_version"`
	Codec          codec.Config     `config:"codec"`
	OltpEndpoint   string           `config:"oltp_endpoint"`
	RetryInterval  int              `config:"retry_interval"`
	Timeout        int              `config:"timeout"`
	BulkMaxSize    int              `config:"bulk_max_size"`
	MaxRetries     int              `config:"max_retries"`
	Queue          config.Namespace `config:"queue"`
	Type           string           `config:"type"`
	TLSCredentials string           `config:"tls_credentials"`
	TLSServerURL   string           `config:"tls_server_url"`
}

// default config object
var (
	defaultConfig = OtlConfig{
		ServiceName:    "sys-devices-vunet",
		ServiceVersion: "1.0.0",
		OltpEndpoint:   "localhost:4317",
		RetryInterval:  60,
		Timeout:        300,
		BulkMaxSize:    1000,
		MaxRetries:     3,
		Type:           "logs",
		TLSCredentials: "",
		TLSServerURL:   "",
	}
)
