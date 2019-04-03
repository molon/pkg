package tracing

import (
	"fmt"
	"strings"

	"github.com/jtolds/gls"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	jaeger "github.com/uber/jaeger-client-go"
)

var mgr = gls.NewContextManager()

type glsSpanKey struct{}

var glsTracingSpanKey = glsSpanKey{}

func GetGlsTracingSpan() opentracing.Span {
	val, ok := mgr.GetValue(glsTracingSpanKey)
	if ok {
		s, ok := val.(opentracing.Span)
		if ok {
			return s
		}
	}
	return nil
}

func SetGlsTracingSpan(sp opentracing.Span, call func()) {
	mgr.SetValues(gls.Values{
		glsTracingSpanKey: sp,
	}, call)
}

// 开始无gls的执行，主要用处是一些特殊情况下断链，防止调用链过长
func SetNonGlsTracingSpan(call func()) {
	mgr.SetValues(gls.Values{
		glsTracingSpanKey: nil,
	}, call)
}

// 下面是基于gls的一些便利方法

func CurrentSpan() opentracing.Span {
	return GetGlsTracingSpan()
}

func CurrentTraceID() string {
	curSp := CurrentSpan()
	if sp, ok := curSp.(*jaeger.Span); ok {
		if spCtx, ok := sp.Context().(jaeger.SpanContext); ok {
			var traceID string
			if spCtx.TraceID().High == 0 {
				traceID = fmt.Sprintf("%x", spCtx.TraceID().Low)
			} else {
				traceID = fmt.Sprintf("%x%016x", spCtx.TraceID().High, spCtx.TraceID().Low)
			}
			return traceID
		}
	}
	return ""
}

func SetTag(key string, value interface{}) opentracing.Span {
	glsSpan := GetGlsTracingSpan()
	if glsSpan != nil {
		return glsSpan.SetTag(key, value)
	}
	return nil
}

func LogError(err error) {
	glsSpan := GetGlsTracingSpan()
	if glsSpan != nil {
		glsSpan.LogFields(ErrorField(err))
		ext.Error.Set(glsSpan, true)
	}
}

func LogFields(fields ...log.Field) {
	glsSpan := GetGlsTracingSpan()
	if glsSpan != nil {
		glsSpan.LogFields(fields...)
	}
}

func Logf(format string, v ...interface{}) {
	format = strings.TrimSuffix(format, "\n")
	LogFields(log.String("event", fmt.Sprintf(format, v...)))
}
