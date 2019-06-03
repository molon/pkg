package tracing

import (
	"context"
	"fmt"

	"github.com/molon/pkg/errors"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// 这俩是比较特殊的key，经常查问题是定位到某个用户的
// 这里我们尽量把用户唯一标识在整个调用链路上传递起来
const BaggageItemKeyUserID = "uid"
const TagKeyUserID = "uid"

// 开始一个span，内部会根据ctx或者gls信息来将链路串起来
func StartSpan(ctx context.Context, call func(ctx context.Context, sp opentracing.Span) error,
	operationName string, component string, opts ...opentracing.StartSpanOption) error {
	if !opentracing.IsGlobalTracerRegistered() {
		return call(ctx, nil)
	}

	if opentracing.SpanFromContext(ctx) == nil {
		//如果ctx里没传，就从gls获取
		glsSpan := GetGlsTracingSpan()
		if glsSpan != nil {
			ctx = opentracing.ContextWithSpan(ctx, glsSpan)
		}
	}

	var parentCtx opentracing.SpanContext
	if parent := opentracing.SpanFromContext(ctx); parent != nil {
		parentCtx = parent.Context()
	}

	opts = append(opts, opentracing.ChildOf(parentCtx))
	opts = append(opts, opentracing.Tag{Key: string(ext.Component), Value: component})

	sp := opentracing.GlobalTracer().StartSpan(
		operationName,
		opts...,
	)
	defer sp.Finish()
	defer func() {
		r := recover() //简单recover记录下，再丢出去
		if r != nil {
			ext.Error.Set(sp, true)
			perr, ok := r.(error)
			if !ok {
				perr = fmt.Errorf(fmt.Sprintln(r))
			}
			sp.LogFields(ErrorField(errors.Wrap(perr, "panic")))

			panic(r)
		}
	}()

	var err error
	SetGlsTracingSpan(sp, func() {
		ctx = opentracing.ContextWithSpan(ctx, sp)
		err = call(ctx, sp)
	})

	//uid
	uid := sp.BaggageItem(BaggageItemKeyUserID)
	if uid != "" {
		sp.SetTag(TagKeyUserID, uid)
	}

	if err != nil {
		ext.Error.Set(sp, true)
		sp.LogFields(ErrorField(err))
	}
	return err
}
