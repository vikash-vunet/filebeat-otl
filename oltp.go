package filebeatotl

import (
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/outputs"
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
	cfg *common.Config,
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
		config.TargetURL,
	)

	if err != nil {
		return outputs.Fail(err)
	}

	logger.Debug("Client Created")

	return outputs.Success(config.BulkMaxSize, config.MaxRetries, client)
}
