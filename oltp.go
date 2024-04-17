package filebeatotl

import (
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/outputs"
	"github.com/elastic/elastic-agent-libs/config"
)

// entrypoint of the plugin
func init() {
	outputs.RegisterType("otlp", makeOtlp)
}

/*
The method to create successful output connection with config and new client
*/
func makeOtlp(
	_ outputs.IndexManager,
	beat beat.Info,
	observer outputs.Observer,
	cfg config.Namespace,
) (outputs.Group, error) {

	logger.Debug("initialize otlp output")

	// config object for OTLP
	config := defaultConfig

	if err := cfg.Unpack(&config); err != nil {
		return outputs.Fail(err)
	}

	logger.Debug("Config loaded")

	// new client object of OTLP
	client, err := newClient(
		observer, config.OltpEndpoint,
		config.ServiceName,
		config.ServiceVersion,
		time.Duration(config.RetryInterval),
	)

	if err != nil {
		return outputs.Fail(err)
	}

	logger.Debug("Client Created")

	return outputs.Success(cfg, config.BulkMaxSize, config.MaxRetries, client)
}
