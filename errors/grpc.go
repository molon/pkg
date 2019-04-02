package errors

import (
	"context"

	"google.golang.org/grpc"
)

func UnaryServerInterceptor(print func(err error)) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil && print != nil {
			print(err)
		}
		return resp, Cause(err)
	}
}

func StreamServerInterceptor(print func(err error)) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, stream)
		if err != nil && print != nil {
			print(err)
		}
		return Cause(err)
	}
}
