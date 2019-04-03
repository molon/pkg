package tracing

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"github.com/uber/jaeger-client-go/config"
)

const collectorEndpointSuffix = "/api/traces?format=jaeger.thrift"

type JaegerLoggerAdapter struct{}

func (l *JaegerLoggerAdapter) Error(msg string) {
	log.Println("Jaeger Error:", msg)
}

func (l *JaegerLoggerAdapter) Infof(msg string, args ...interface{}) {
	log.Println("Jaeger:", fmt.Sprintf(msg, args...))
}

type logrusWrapper struct {
	*logrus.Logger
}

func (l *logrusWrapper) Error(msg string) {
	l.Logger.Errorf(msg)
}

type JaegerTracer struct {
	opentracing.Tracer
	closer io.Closer
}

func (t *JaegerTracer) Close() error {
	return t.closer.Close()
}

func DefaultJaegerConfiguration(svc string, collectorEndpoint string) config.Configuration {
	svc = strings.Replace(svc, "://", "_", -1)
	svc = strings.Replace(svc, ":", "_", -1)
	svc = strings.Replace(svc, "/", "_", -1)

	if !strings.HasSuffix(collectorEndpoint, collectorEndpointSuffix) {
		collectorEndpoint = strings.TrimSuffix(collectorEndpoint, "/") + collectorEndpointSuffix
	}
	if !strings.HasPrefix(collectorEndpoint, "http://") {
		collectorEndpoint = "http://" + collectorEndpoint
	}

	return config.Configuration{
		ServiceName: svc,
		//直接全都记录，忽略采样，我们这里是当做日志来记
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			// 是否打印log，这个只是不打印Info，但是Error还是会打印的
			LogSpans:            false,
			BufferFlushInterval: 2 * time.Second,
			QueueSize:           3000,
			CollectorEndpoint:   collectorEndpoint,
		},
	}
}

func NewJaegerTracer(logger *logrus.Logger, cfg config.Configuration, options ...config.Option) (*JaegerTracer, error) {
	opts := []config.Option{
		config.ZipkinSharedRPCSpan(false),
	}
	if logger != nil {
		opts = append(opts, config.Logger(&logrusWrapper{Logger: logger}))
	}

	opts = append(opts, options...)

	t, closer, err := cfg.NewTracer(opts...)
	if err != nil {
		return nil, err
	}

	opentracing.SetGlobalTracer(t)

	return &JaegerTracer{
		Tracer: t,
		closer: closer,
	}, nil
}
