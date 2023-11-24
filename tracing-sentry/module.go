package tracing_sentry

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/getsentry/sentry-go"

	"go.uber.org/zap"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	caddy.RegisterModule(Tracing{})
	httpcaddyfile.RegisterHandlerDirective("sentry", parseCaddyfile)
	// 初始化sentry
	sampleRate, _ := strconv.ParseFloat(os.Getenv("SENTRY_SAMPLE_RATE"), 64)
	if sampleRate == 0.0 {
		sampleRate = 1.0
	}
	serverName := os.Getenv("SENTRY_SERVICE_NAME")
	if serverName == "" {
		serverName = os.Getenv("OTEL_SERVICE_NAME")
	}
	if serverName == "" {
		serverName = os.Getenv("SERVICE_NAME")
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn:                "",
		EnableTracing:      true,
		SampleRate:         sampleRate,
		TracesSampleRate:   sampleRate,
		ProfilesSampleRate: sampleRate,
		Debug:              os.Getenv("SENTRY_DEBUG") == "true",
		ServerName:         serverName, // 默认hostname
		//TracesSampler: sentry.TracesSampler(func(ctx sentry.SamplingContext) float64 {
		//	// Don't sample health checks.
		//	if ctx.Span.Name == "GET /health" {
		//		return 0.0
		//	}
		//
		//	return 1.0
		//}),
	})
	if err != nil {
		fmt.Println("sentry init error", err)
	}

}

// Tracing implements an HTTP handler that adds support for distributed tracing,
// using OpenTelemetry. This module is responsible for the injection and
// propagation of the trace context. Configure this module via environment
// variables (see https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/sdk-environment-variables.md).
// Some values can be overwritten in the configuration file.
type Tracing struct {
	// SpanName is a span name. It should follow the naming guidelines here:
	// https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/api.md#span
	SpanName string `json:"span"`

	// otel implements opentelemetry related logic.
	otel openTelemetryWrapper

	logger *zap.Logger
}

// CaddyModule returns the Caddy module information.
func (Tracing) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.sentry",
		New: func() caddy.Module { return new(Tracing) },
	}
}

// Provision implements caddy.Provisioner.
func (ot *Tracing) Provision(ctx caddy.Context) error {
	ot.logger = ctx.Logger()

	var err error
	ot.otel, err = newOpenTelemetryWrapper(ctx, ot.SpanName)

	return err
}

// Cleanup implements caddy.CleanerUpper and closes any idle connections. It
// calls Shutdown method for a trace provider https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/sdk.md#shutdown.
func (ot *Tracing) Cleanup() error {
	if err := ot.otel.cleanup(ot.logger); err != nil {
		return fmt.Errorf("tracerProvider shutdown: %w", err)
	}
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (ot *Tracing) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	return ot.otel.ServeHTTP(w, r, next)
}

// UnmarshalCaddyfile sets up the module from Caddyfile tokens. Syntax:
//
//	tracing {
//	    [span <span_name>]
//	}
func (ot *Tracing) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	setParameter := func(d *caddyfile.Dispenser, val *string) error {
		if d.NextArg() {
			*val = d.Val()
		} else {
			return d.ArgErr()
		}
		if d.NextArg() {
			return d.ArgErr()
		}
		return nil
	}

	// paramsMap is a mapping between "string" parameter from the Caddyfile and its destination within the module
	paramsMap := map[string]*string{
		"span": &ot.SpanName,
	}

	for d.Next() {
		args := d.RemainingArgs()
		if len(args) > 0 {
			return d.ArgErr()
		}

		for d.NextBlock(0) {
			if dst, ok := paramsMap[d.Val()]; ok {
				if err := setParameter(d, dst); err != nil {
					return err
				}
			} else {
				return d.ArgErr()
			}
		}
	}
	return nil
}

func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m Tracing
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return &m, err
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Tracing)(nil)
	_ caddyhttp.MiddlewareHandler = (*Tracing)(nil)
	_ caddyfile.Unmarshaler       = (*Tracing)(nil)
)
