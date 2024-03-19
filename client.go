package filebeatotl

import (
	// "context"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/libbeat/outputs"
	"github.com/elastic/beats/v7/libbeat/outputs/codec"
	"github.com/elastic/beats/v7/libbeat/publisher"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
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
	tp 		   *sdktrace.TracerProvider
	tracer 	   trace.Tracer
	targetURL      string
}

func newClient(observer outputs.Observer, endpoint string, service_name string, service_version string, timeout time.Duration, targetURL string) (*client, error) {
	c := &client{
		log:          logp.NewLogger("otlp"),
		observer:     observer,
		oltpEndpoint: endpoint,
		timeout:      timeout,
		serviceName: service_name,
		serviceVersion: service_version,
		targetURL: targetURL,
	}

	return c, nil
}

func newExporter(ctx context.Context, c *client) (*otlptrace.Exporter, error) {
	var headers = map[string]string{
		// Remove the Lightstep token handling
	}

	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithHeaders(headers),
		otlptracegrpc.WithEndpoint(c.oltpEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	return otlptrace.New(ctx, client)
}

func newTraceProvider(exp *otlptrace.Exporter, c *client) *sdktrace.TracerProvider {
	if len(c.serviceName) == 0 {
		c.serviceName = "sys-devices-pci0000:00-0000:00:03.0-net-enp0s3.device"
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
	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource),
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

	tp := newTraceProvider(exp, c)
	logger.Debug("new trace provider set up")
	c.tp = tp
	otel.SetTracerProvider(tp)
	logger.Debug("set trace provider")

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	c.tracer = tp.Tracer(c.serviceName, trace.WithInstrumentationVersion(c.serviceVersion))
	fmt.Println(c.tracer)

	c.ctx = ctx;

	logger.Debug("connection successful")

	return nil
}

func (c *client) Close() error {
	// Implement closing logic
	func() { _ = c.tp.Shutdown(c.ctx) }()
	logger.Debug("closed connection")
	return nil
}

func (c *client) Publish(ctx context.Context, batch publisher.Batch) error {
	// Implement publishing logic

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

	for _ , event := range events {
		content := event.Content
		data, err := content.GetValue("message")
		if err != nil {
			fmt.Println("Error getting value from event")
		}

		mergedData := map[string]interface{}{
			"data": 12,
			"message": data,
		}
			jsonData, err := json.Marshal(mergedData)
			if err != nil {
				fmt.Printf("Error encoding data to JSON: %v\n", err)
				return err
			} else {
				makeRequest(c.ctx, jsonData, c)
			}

	}

	return nil
}

func makeRequest(ctx context.Context, jsonData []byte, c *client) {
	// Start a span for the HTTP request
	logger.Debug("started requests")
	ctx, span := c.tracer.Start(ctx, "makeRequest")
	defer span.End()

	span.AddEvent("Did some cool stuff")

	// Marshal uptime data to JSON

	// Create a new HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.targetURL, strings.NewReader(string(jsonData)))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}

	// Inject the span context into the HTTP request
	propagation.TraceContext{}.Inject(ctx, propagation.HeaderCarrier(req.Header))

	// Use a standard HTTP client to perform the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request: %v", err)
		return
	}
	defer res.Body.Close()

	// Print request details
	fmt.Printf("Request to %s\n", c.targetURL)
	fmt.Printf("Status Code: %d\n", res.StatusCode)
	fmt.Printf("Headers: %v\n", res.Header)

	// Don't read the response body here
	// ...

	// Set attributes for the span
	span.SetAttributes(
		attribute.Int("status_code", res.StatusCode),
	)

	span.AddEvent("Cancelled wait due to external signal", trace.WithAttributes(attribute.Int("pid", 4328), attribute.String("signal", "SIGHUP")))
}

func (c *client) String() string {
	return "otlp(" + c.oltpEndpoint + ")"
}

// Implement other necessary methods and helper functions here
