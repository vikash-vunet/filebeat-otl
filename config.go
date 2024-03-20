package filebeatotl

import (
	"github.com/elastic/beats/v7/libbeat/outputs/codec"
)

// config object structure to store all OTLP Configs
type OtlConfig struct {
	ServiceName    string       `config:"service_name"`
	ServiceVersion string       `config:"service_version"`
	Codec          codec.Config `config:"codec"`
	TargetURL      string       `config:"target_url"`
	OltpEndpoint   string       `config:"oltp_endpoint"`
	RetryInterval  int          `config:"retry_interval"`
	Timeout        int          `config:"timeout"`
	BulkMaxSize    int          `config:"bulk_max_size"`
	MaxRetries     int          `config:"max_retries"`
}

// default config object
var (
	defaultConfig = OtlConfig{
		ServiceName:    "sys-devices-vunet",
		ServiceVersion: "1.0.0",
		TargetURL:      "http://localhost:8081/uptime",
		OltpEndpoint:   "localhost:4317",
		RetryInterval:  60,
		Timeout:        300,
		BulkMaxSize:    1000,
		MaxRetries:     3,
	}
)
