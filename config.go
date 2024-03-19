package filebeatotl

import (
	"github.com/elastic/beats/v7/libbeat/outputs/codec"
)

type OtlConfig struct {
	serviceName    string       `config:"service_name"`
	serviceVersion string       `config:"service_version"`
	codec          codec.Config `config:"codec"`
	targetURL      string       `config:"target_url"`
	oltpEndpoint   string       `config:"oltp_endpoint"`
	retryInterval  int          `config:"retry_interval"`
	timeout        int          `config:"timeout"`
	BulkMaxSize    int          `config:"bulk_max_size"`
	MaxRetries     int          `config:"max_retries"`
}

var (
	defaultConfig = OtlConfig{
		serviceName:    "sys-devices-pci0000:00-0000:00:03.0-net-enp0s3.device",
		serviceVersion: "1.0.0",
		targetURL:      "http://localhost:8081/uptime",
		oltpEndpoint:   "localhost:4317",
		retryInterval:  60,
		timeout:        300,
		BulkMaxSize:    1000,
		MaxRetries:     3,
	}
)
