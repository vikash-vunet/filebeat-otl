package filebeatotl

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	// "net/http"
	// "strings"
	"time"

	"github.com/agoda-com/opentelemetry-logs-go"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs/otlplogsgrpc"
	"github.com/agoda-com/opentelemetry-logs-go/logs"
	sdk "github.com/agoda-com/opentelemetry-logs-go/sdk/logs"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/beats/v7/libbeat/outputs"
	"github.com/elastic/beats/v7/libbeat/outputs/codec"
	"github.com/elastic/beats/v7/libbeat/publisher"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
	"google.golang.org/grpc/credentials"
)

var (
	logger = logp.NewLogger("otlp")
)

type client struct {
	log            *logp.Logger
	observer       outputs.Observer
	oltpEndpoint   string
	timeout        time.Duration
	serviceName    string
	serviceVersion string
	ctx            context.Context
	codec          codec.Codec
	lp             *sdk.LoggerProvider
	logEmitter     logs.Logger
}

func newClient(
	observer outputs.Observer,
	endpoint string,
	service_name string,
	service_version string,
	timeout time.Duration,
) (*client, error) {
	c := &client{
		log:            logp.NewLogger("otlp"),
		observer:       observer,
		oltpEndpoint:   endpoint,
		timeout:        timeout,
		serviceName:    service_name,
		serviceVersion: service_version,
	}

	return c, nil
}

func newExporter(ctx context.Context, c *client) (*otlplogs.Exporter, error) {
	var headers = map[string]string{
		// Remove the Lightstep token handling
	}

	creds, err := credentials.NewClientTLSFromFile("/home/vunet-systems/development/otel_cert/cacert.pem", "127.0.0.1")
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS credentials: %v", err)
	}

	client := otlplogsgrpc.NewClient(
		otlplogsgrpc.WithHeaders(headers),
		otlplogsgrpc.WithEndpoint(c.oltpEndpoint),
		otlplogsgrpc.WithTLSCredentials(creds),
	)

	return otlplogs.NewExporter(ctx, otlplogs.WithClient(client))
}

func newLogProvider(exp *otlplogs.Exporter, c *client) *sdk.LoggerProvider {
	if len(c.serviceName) == 0 {
		c.serviceName = "sys-devices-vunet"
		log.Printf("Using default service name %s", c.serviceName)
	}

	if len(c.serviceVersion) == 0 {
		c.serviceVersion = "0.1.0"
		log.Printf("Using default service version %s", c.serviceVersion)
	}

	resource, rErr :=
		resource.Merge(
			resource.Default(),
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(c.serviceName),
				semconv.ServiceVersionKey.String(c.serviceVersion),
				attribute.String("environment", "dev"),
			),
		)

	if rErr != nil {
		panic(rErr)
	}

	return sdk.NewLoggerProvider(
		sdk.WithBatcher(exp),
		sdk.WithResource(resource),
	)
}

func (c *client) Connect() error {
	// Implement connection logic
	ctx := context.Background()

	logger.Debug("connection started")
	exp, err := newExporter(ctx, c)
	if err != nil {
		log.Fatalf("failed to initialize exporter: %v", err)
	}

	c.lp = newLogProvider(exp, c)
	logger.Debug("new trace provider set up")

	otel.SetLoggerProvider(c.lp)
	logger.Debug("set trace provider")

	c.logEmitter = otel.GetLoggerProvider().Logger(
		"otel",
		logs.WithInstrumentationVersion("0.0.1"),
		logs.WithSchemaURL(semconv.SchemaURL),
	)

	c.ctx = ctx

	logger.Debug("connection successful")

	return nil
}

func (c *client) Close() error {
	// Implement closing logic
	func() { _ = c.lp.Shutdown(c.ctx) }()
	logger.Debug("closed connection")
	return nil
}

func (c *client) Publish(ctx context.Context, batch publisher.Batch) error {
	// Implement publishing logic
	fmt.Println("time started", time.Now().Local())
	if c == nil {
		panic("no client")
	}
	if batch == nil {
		panic("no batch")
	}

	logger.Debug("publish started")

	events := batch.Events()
	c.observer.NewBatch(len(events))
	logger.Debug("Started reading events")

	var retryEvent []publisher.Event

	for _, event := range events {
		content := event.Content
		data, err := content.GetValue("message")
		if err != nil {
			retryEvent = append(retryEvent, event)
			fmt.Println("Error getting value from event")
		}

		// mergedData := map[string]interface{}{
		// 	"log": data,
		// }
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Printf("Error encoding data to JSON: %v\n", err)
			retryEvent = append(retryEvent, event)
		} else {
			go makeRequest(jsonData, c)
		}
	}
	if len(retryEvent) != 0 {
		batch.RetryEvents(retryEvent)
	} else {
		batch.ACK()
	}

	fmt.Println("time ended", time.Now().Local())

	return nil
}

func makeRequest(jsonData []byte, c *client) {
	// Start a span for the HTTP request
	logger.Debug("started requests")

	s := string(jsonData)
	lrc := logs.LogRecordConfig{
		Timestamp: nil,
		ObservedTimestamp: time.Now(),
		Body: &s,
	}
	logRecord := logs.NewLogRecord(lrc)
	c.logEmitter.Emit(logRecord)

}

func (c *client) String() string {
	return "otlp(" + c.oltpEndpoint + ")"
}
