package otgrpc

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/empty"
	eco "github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	jsoniter "github.com/json-iterator/go"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/molon/pkg/tracing"
)

type metadataReaderWriter struct {
	metadata.MD
}

func (w metadataReaderWriter) Set(key, val string) {
	key = strings.ToLower(key)
	w.MD[key] = append(w.MD[key], val)
}

func (w metadataReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range w.MD {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}

	return nil
}

// UnaryServerInterceptor
func UnaryServerInterceptor(options ...TracingOption) grpc.UnaryServerInterceptor {
	tOpts := &tracingOptions{}
	for _, opt := range options {
		opt(tOpts)
	}

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		if !opentracing.IsGlobalTracerRegistered() {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		spanCtx, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, metadataReaderWriter{md})

		var op string
		if tOpts.opNameFunc != nil {
			op = tOpts.opNameFunc()
		}
		if op == "" {
			op = fmt.Sprintf("GRPC %s", info.FullMethod)
		}
		sp := opentracing.GlobalTracer().StartSpan(
			op,
			ext.RPCServerOption(spanCtx),
		)
		defer sp.Finish()
		defer func() {
			r := recover() //简单recover记录下，再丢出去
			if r != nil {
				ext.Error.Set(sp, true)
				perr, ok := r.(error)
				if !ok {
					perr = fmt.Errorf(fmt.Sprint(r))
				}
				sp.LogFields(tracing.ErrorField(errors.Wrap(perr, "panic")))

				panic(r)
			}
		}()

		//设置tag
		ext.Component.Set(sp, "grpc")

		//记录请求
		traceMD(sp, md)
		if tOpts.requestBody {
			traceRequest(sp, req, tOpts.maxBodyLogSize)
		}

		//执行请求，gls包裹，这样interceptor内部的调用都会从gls自动获取当前调用链
		tracing.SetGlsTracingSpan(sp, func() {
			ctx = opentracing.ContextWithSpan(ctx, sp)
			resp, err = handler(ctx, req)
		})

		//uid
		uid := sp.BaggageItem(tracing.BaggageItemKeyUserID)
		if uid != "" {
			sp.SetTag(tracing.TagKeyUserID, uid)
		}

		//记录错误
		if err != nil {
			eco.SetSpanTags(sp, err, false)       //这个里面能设置下错误code和class，grpc的标准
			ext.Error.Set(sp, true)               //上个方法里并没有设置
			sp.LogFields(tracing.ErrorField(err)) //server端不需要在这里的堆栈信息
		}

		//记录resp
		if tOpts.responseBody {
			traceResponse(sp, resp, tOpts.maxBodyLogSize)
		}
		return
	}
}

// UnaryClientInterceptor ...
func UnaryClientInterceptor(options ...TracingOption) grpc.UnaryClientInterceptor {
	tOpts := &tracingOptions{}
	for _, opt := range options {
		opt(tOpts)
	}

	return func(
		ctx context.Context,
		method string,
		req, resp interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) (err error) {
		if !opentracing.IsGlobalTracerRegistered() {
			return invoker(ctx, method, req, resp, cc, opts...)
		}

		if opentracing.SpanFromContext(ctx) == nil {
			//如果ctx里没传，就从gls获取
			glsSpan := tracing.GetGlsTracingSpan()
			if glsSpan != nil {
				ctx = opentracing.ContextWithSpan(ctx, glsSpan)
			}
		}

		var parentCtx opentracing.SpanContext
		if parent := opentracing.SpanFromContext(ctx); parent != nil {
			parentCtx = parent.Context()
		}

		var op string
		if tOpts.opNameFunc != nil {
			op = tOpts.opNameFunc()
		}
		if op == "" {
			op = fmt.Sprintf("GRPC_CLI %s", method)
		}
		sp := opentracing.GlobalTracer().StartSpan(
			op,
			opentracing.ChildOf(parentCtx),
			ext.SpanKindRPCClient,
		)
		defer sp.Finish()
		defer func() {
			r := recover() //简单recover记录下，再丢出去
			if r != nil {
				ext.Error.Set(sp, true)
				perr, ok := r.(error)
				if !ok {
					perr = fmt.Errorf(fmt.Sprint(r))
				}
				sp.LogFields(tracing.ErrorField(errors.Wrap(perr, "panic")))

				panic(r)
			}
		}()

		//设置tag
		ext.Component.Set(sp, "grpc")

		//设置carrier
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}
		mdWriter := metadataReaderWriter{md}
		err = sp.Tracer().Inject(sp.Context(), opentracing.HTTPHeaders, mdWriter)
		if err != nil {
			ext.Error.Set(sp, true)
			sp.LogFields(tracing.ErrorField(errors.Wrap(err, "Tracer.Inject() failed")))
		}
		ctx = metadata.NewOutgoingContext(ctx, md)

		//记录请求
		traceMD(sp, md)
		if tOpts.requestBody {
			traceRequest(sp, req, tOpts.maxBodyLogSize)
		}

		//执行请求
		err = invoker(ctx, method, req, resp, cc, opts...)

		//uid
		uid := sp.BaggageItem(tracing.BaggageItemKeyUserID)
		if uid != "" {
			sp.SetTag(tracing.TagKeyUserID, uid)
		}

		//记录错误
		if err != nil {
			eco.SetSpanTags(sp, err, true) //这个里面能设置下错误code和class，grpc的标准
			// ext.Error.Set(sp, true) //上个语句已经设置
			sp.LogFields(tracing.ErrorField(errors.Wrap(err, "Invoke failed")))
		}

		//记录resp
		if tOpts.responseBody {
			traceResponse(sp, resp, tOpts.maxBodyLogSize)
		}

		return
	}
}

func traceRequest(sp opentracing.Span, req interface{}, maxBodyLogSize int) {
	jsn, err := marshalBodyToString(req)
	if err != nil {
		ext.Error.Set(sp, true)
		sp.LogFields(tracing.ErrorField(errors.Wrap(err, "Marshal request.body failed")))
	} else {
		sp.LogFields(log.String("request.body", tracing.PruneBodyLog(jsn, maxBodyLogSize)))
	}
}

func traceResponse(sp opentracing.Span, resp interface{}, maxBodyLogSize int) {
	jsn, err := marshalBodyToString(resp)
	if err != nil {
		ext.Error.Set(sp, true)
		sp.LogFields(tracing.ErrorField(errors.Wrap(err, "Marshal response.body failed")))
	} else {
		sp.LogFields(log.String("response.body", tracing.PruneBodyLog(jsn, maxBodyLogSize)))
	}
}

func traceMD(sp opentracing.Span, md metadata.MD) {
	jsn, err := jsoniter.Marshal(md)
	if err != nil {
		ext.Error.Set(sp, true)
		sp.LogFields(tracing.ErrorField(errors.Wrap(err, "Marshal metadata failed")))
	} else {
		sp.LogFields(log.String("metadata", string(jsn)))
	}
}

func marshalBodyToString(b interface{}) (string, error) {
	pb, ok := b.(proto.Message)
	if ok {
		marshaler := &jsonpb.Marshaler{
			EmitDefaults: true,
			OrigName:     true,
			AnyResolver:  anyResolver{},
		}
		return marshaler.MarshalToString(pb)
	}

	return jsoniter.MarshalToString(b)
}

type anyResolver struct{}

func (anyResolver) Resolve(typeUrl string) (proto.Message, error) {
	mname := typeUrl
	if slash := strings.LastIndex(mname, "/"); slash >= 0 {
		mname = mname[slash+1:]
	}
	mt := proto.MessageType(mname)
	if mt == nil {
		return &empty.Empty{}, nil
	}
	return reflect.New(mt.Elem()).Interface().(proto.Message), nil
}
