package filebeatotl

import (
	"context"
	// "encoding/json"
	"fmt"
	"log"

	// "net/http"
	// "strings"
	"time"

	"github.com/agoda-com/opentelemetry-logs-go"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs/otlplogsgrpc"
	"github.com/agoda-com/opentelemetry-logs-go/logs"
	logssdk "github.com/agoda-com/opentelemetry-logs-go/sdk/logs"
	"github.com/elastic/beats/v7/libbeat/outputs"
	"github.com/elastic/beats/v7/libbeat/outputs/codec"
	"github.com/elastic/beats/v7/libbeat/publisher"
	"github.com/elastic/elastic-agent-libs/logp"
	otelMetric "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
	"google.golang.org/grpc/credentials"
)

var (
	logger = logp.NewLogger("otlp")
)

// Client Connection details struct having all configs
type client struct {
	log                   *logp.Logger
	observer              outputs.Observer
	oltpEndpoint          string
	timeout               time.Duration
	serviceName           string
	serviceVersion        string
	serviceType           string
	serviceTLSCredentials string
	serviceTLSServerURL   string
	ctx                   context.Context
	codec                 codec.Codec
	lp                    *logssdk.LoggerProvider
	mp                    *metricsdk.MeterProvider
	logEmitter            logs.Logger
	meter                 metric.Float64Counter
	index                 string
}

// create a new client with config
func newClient(
	observer outputs.Observer,
	endpoint string,
	service_name string,
	service_version string,
	timeout time.Duration,
	index string,
	codec codec.Codec,
	serviceType string,
	serviceTLSCredentials string,
	serviceTLSServerURL string,

) (*client, error) {
	c := &client{
		log:                   logp.NewLogger("otlp"),
		observer:              observer,
		oltpEndpoint:          endpoint,
		timeout:               timeout,
		serviceName:           service_name,
		serviceVersion:        service_version,
		index:                 index,
		codec:                 codec,
		serviceType:           serviceType,
		serviceTLSCredentials: serviceTLSCredentials,
		serviceTLSServerURL:   serviceTLSServerURL,
	}

	return c, nil
}

// A new log exporter function with tls certification and header options
func newLogsExporter(ctx context.Context, c *client) (*otlplogs.Exporter, error) {
	var headers = map[string]string{
		// Remove the Lightstep token handling
	}

	if c.serviceTLSCredentials == "" || c.serviceTLSServerURL == "" {
		client := otlplogsgrpc.NewClient(
			otlplogsgrpc.WithHeaders(headers),
			otlplogsgrpc.WithEndpoint(c.oltpEndpoint),
		)
	
		return otlplogs.NewExporter(ctx, otlplogs.WithClient(client))
	}


	creds, err := credentials.NewClientTLSFromFile(c.serviceTLSCredentials, c.serviceTLSServerURL)
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

// A new metric exporter function with tls certification and header options
func newMetricsExporter(ctx context.Context, c *client) (*otlpmetricgrpc.Exporter, error) {
	var headers = map[string]string{
		// Remove the Lightstep token handling
	}

	if c.serviceTLSCredentials == "" || c.serviceTLSServerURL == "" {
		return otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithHeaders(headers),
		otlpmetricgrpc.WithEndpoint(c.oltpEndpoint))
	}

	creds, err := credentials.NewClientTLSFromFile(c.serviceTLSCredentials, c.serviceTLSServerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS credentials: %v", err)
	}

	return otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithHeaders(headers),
		otlpmetricgrpc.WithEndpoint(c.oltpEndpoint),
		otlpmetricgrpc.WithTLSCredentials(creds))
}

// A new log provider function
func newLogProvider(exp *otlplogs.Exporter, c *client) *logssdk.LoggerProvider {
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
			),
		)

	if rErr != nil {
		panic(rErr)
	}

	return logssdk.NewLoggerProvider(
		logssdk.WithBatcher(exp),
		logssdk.WithResource(resource),
	)
}

// A new metric Provider function
func newMetricProvider(exp *otlpmetricgrpc.Exporter, c *client) *metricsdk.MeterProvider {
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

	return metricsdk.NewMeterProvider(
		metricsdk.WithReader(metricsdk.NewPeriodicReader(exp)),
		metricsdk.WithResource(resource),
	)
}

func (c *client) Connect() error {
	// Implement connection logic
	ctx := context.Background()
	logger.Debug("connection started")

	if c.serviceType == "logs" {
		exp, err := newLogsExporter(ctx, c)
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

	} else if c.serviceType == "metrics" {
		exp, err := newMetricsExporter(ctx, c)
		if err != nil {
			log.Fatalf("failed to initialize exporter: %v", err)
		}

		c.mp = newMetricProvider(exp, c)
		logger.Debug("new metric provider set up")

		otelMetric.SetMeterProvider(c.mp)
		logger.Debug("set trace provider")

		c.meter, _ = otelMetric.Meter(
			"otel-meter",
			metric.WithSchemaURL(semconv.SchemaURL),
			metric.WithInstrumentationVersion("0.0.1"),
		).Float64Counter(
			"demo_server/request_counts",
			metric.WithDescription("The number of requests received"),
		)
	} else {
		return fmt.Errorf("Invalid type found")
	}

	c.ctx = ctx

	logger.Debug("connection successful")

	return nil
}

func (c *client) Close() error {
	// Implement closing logic
	if c.serviceType == "logs" {
		func() { _ = c.lp.Shutdown(c.ctx) }()
	} else if c.serviceType == "metrics" {
		func() { _ = c.mp.Shutdown(c.ctx) }()
	} else {
		return fmt.Errorf("Invalid type found")
	}
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
		content := &event.Content
		//  attr :=  &[]attribute.KeyValue{}

		// message, err := content.GetValue("message")
		// if err != nil {
		// 	retryEvent = append(retryEvent, event)
		// 	fmt.Println("Error getting message from event")
		// }

		data, err := c.codec.Encode(c.index, content)

		// jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Printf("Error encoding data to JSON: %v\n", err)
			retryEvent = append(retryEvent, event)
		} else {
			makeRequest(data, c)
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
	fmt.Println("output ", s)
	fmt.Println("---")

	if c.serviceType == "logs" {
		lrc := logs.LogRecordConfig{
			Timestamp:         nil,
			ObservedTimestamp: time.Now(),
			Body:              &s,
		}
		logRecord := logs.NewLogRecord(lrc)
		c.logEmitter.Emit(logRecord)
	} else if c.serviceType == "metrics" {
		requestCount := c.meter
		requestCount.Add(c.ctx, 1, metric.WithAttributes(attribute.String("data", s)))
	}

}

func (c *client) String() string {
	return "otlp(" + c.oltpEndpoint + ")"
}
