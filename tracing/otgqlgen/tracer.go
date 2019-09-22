package otgqlgen

import (
	"context"
	"fmt"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
)

var _ graphql.Tracer = (tracerImpl)(0)

func NewTracer() graphql.Tracer {
	return tracerImpl(0)
}

type tracerImpl int

func (tracerImpl) StartOperationParsing(ctx context.Context) context.Context {
	return ctx
}

func (tracerImpl) EndOperationParsing(ctx context.Context) {
}

func (tracerImpl) StartOperationValidation(ctx context.Context) context.Context {
	return ctx
}

func (tracerImpl) EndOperationValidation(ctx context.Context) {
}

func (tracerImpl) StartOperationExecution(ctx context.Context) context.Context {
	requestContext := graphql.GetRequestContext(ctx)

	// 内省的请求直接过滤掉不记录
	if strings.HasPrefix(requestContext.RawQuery, "query IntrospectionQuery") {
		return ctx
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "graphql")
	span.LogFields(log.String("query", requestContext.RawQuery))
	ext.SpanKind.Set(span, "server")
	ext.Component.Set(span, "gqlgen")

	return ctx
}

func (tracerImpl) StartFieldExecution(ctx context.Context, field graphql.CollectedField) context.Context {
	span, ctx := opentracing.StartSpanFromContext(ctx, "unnamed")
	ext.SpanKind.Set(span, "server")
	ext.Component.Set(span, "gqlgen")

	return ctx
}

func (tracerImpl) StartFieldResolverExecution(ctx context.Context, rc *graphql.ResolverContext) context.Context {
	span := opentracing.SpanFromContext(ctx)
	span.SetOperationName(rc.Object + "_" + rc.Field.Name)
	span.SetTag("resolver.object", rc.Object)
	span.SetTag("resolver.field", rc.Field.Name)

	return ctx
}

func (tracerImpl) StartFieldChildExecution(ctx context.Context) context.Context {
	return ctx
}

func (tracerImpl) EndFieldExecution(ctx context.Context) {
	span := opentracing.SpanFromContext(ctx)
	defer span.Finish()

	rc := graphql.GetResolverContext(ctx)
	reqCtx := graphql.GetRequestContext(ctx)

	errList := reqCtx.GetErrors(rc)
	if len(errList) != 0 {
		ext.Error.Set(span, true)
		span.LogFields(
			log.String("event", "error"),
		)

		for idx, err := range errList {
			span.LogFields(
				log.String(fmt.Sprintf("error.%d.message", idx), err.Error()),
				log.String(fmt.Sprintf("error.%d.kind", idx), fmt.Sprintf("%T", err)),
			)
		}
	}
}

func (tracerImpl) EndOperationExecution(ctx context.Context) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		span.Finish()
	}
}
