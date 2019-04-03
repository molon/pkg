package otgrpc

type tracingOptions struct {
	opNameFunc func() string

	maxBodyLogSize int
	requestBody    bool
	responseBody   bool
}

type TracingOption func(*tracingOptions)

func WithOperationNameFunc(opNameFunc func() string) TracingOption {
	return func(options *tracingOptions) {
		options.opNameFunc = opNameFunc
	}
}

func WithRequstBody(requestBody bool) TracingOption {
	return func(options *tracingOptions) {
		options.requestBody = requestBody
	}
}

func WithResponseBody(responseBody bool) TracingOption {
	return func(options *tracingOptions) {
		options.responseBody = responseBody
	}
}

func WithMaxBodyLogSize(maxBodyLogSize int) TracingOption {
	return func(options *tracingOptions) {
		options.maxBodyLogSize = maxBodyLogSize
	}
}
