


## caddy sentry

https://docs.sentry.io/platforms/go/migration/
```azure
sentry.CaptureException(err)

func() {
	defer sentry.Recover()
	// do all of the scary things here
}()

```
SDK中的环境变量
- SENTRY_DSN 必填
- SENTRY_ENVIRONMENT
- SENTRYGODEBUG 如:httptrace=1,httpdump=1

- caddy sentry中的环境变量
- SENTRY_DEBUG 默认false
- SENTRY_SAMPLE_RATE 默认1.0, 范围(0.0,1.0], 如果全不采集那么DSN为空
- SENTRY_SERVICE_NAME 默认hostname

## caddy tracing

docker-compose.yaml
```yaml
version: '3.7'

services:
  caddy:
    image: caddy:latest
    ports:
      - "80:80"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
    environment:
      - OTEL_SERVICE_NAME=caddy-app
      - OTEL_EXPORTER_OTLP_ENDPOINT=http://<jaeger-ip>:4317
      - OTEL_EXPORTER_OTLP_INSECURE=true
      - OTEL_EXPORTER_OTLP_PROTOCOL=grpc
      - GRPC_GO_LOG_VERBOSITY_LEVEL=99
      - GRPC_GO_LOG_SEVERITY_LEVEL=info

```

[OTLP Exporter Configuration](https://opentelemetry.io/docs/concepts/sdk-configuration/otlp-exporter-configuration/)