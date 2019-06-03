package errors

import (
	"context"

	"google.golang.org/grpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type interceptorOptions struct {
	errHandler func(err error) error
}

type InterceptorOption func(*interceptorOptions)

func WithErrorHandler(errHandler func(err error) error) InterceptorOption {
	return func(options *interceptorOptions) {
		options.errHandler = errHandler
	}
}

func UnaryServerInterceptor(options ...InterceptorOption) grpc.UnaryServerInterceptor {
	opts := &interceptorOptions{}
	for _, option := range options {
		option(opts)
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if opts.errHandler != nil {
			err = opts.errHandler(err)
		}
		return resp, Cause(err)
	}
}

func StreamServerInterceptor(options ...InterceptorOption) grpc.StreamServerInterceptor {
	opts := &interceptorOptions{}
	for _, option := range options {
		option(opts)
	}

	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, stream)
		if opts.errHandler != nil {
			err = opts.errHandler(err)
		}
		return Cause(err)
	}
}

func Statusf(c codes.Code, format string, a ...interface{}) error {
	return WithStack(status.Errorf(c, format, a...))
}
