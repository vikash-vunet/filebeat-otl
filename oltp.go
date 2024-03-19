package filebeatotl

import (
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/libbeat/outputs"
)

// var logger = logp.NewLogger("ClickHouse")

func init() {
	outputs.RegisterType("otlp", makeOtlp)
}

func makeOtlp(
	_ outputs.IndexManager,
	beat beat.Info,
	observer outputs.Observer,
	cfg *common.Config,
) (outputs.Group, error) {

	log := logp.NewLogger("otlp")
	log.Debug("initialize otlp output")

	config := defaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return outputs.Fail(err)
	}

	client, err := newClient(observer, config.oltpEndpoint, config.serviceName, config.serviceVersion, time.Duration(config.retryInterval), config.targetURL)
	if err != nil {
		return outputs.Fail(err)
	}

	retry := 0
	if config.MaxRetries < 0 {
		retry = -1
	}

	return outputs.Success(config.BulkMaxSize, retry, client)
}
