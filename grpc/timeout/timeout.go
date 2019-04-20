package timeout

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

// If ctx without timeout, use the `defaultTimeout`
func UnaryClientInterceptorWhenCall(defaultTimeout time.Duration) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, resp interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if _, ok := ctx.Deadline(); ok {
			return invoker(ctx, method, req, resp, cc, opts...)
		}

		ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
		return invoker(ctx, method, req, resp, cc, opts...)
	}
}
