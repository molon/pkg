package otgorm

type tracingOptions struct {
	opNameFunc func() string

	maxBodyLogSize int
	resultBody     bool
}

type TracingOption func(*tracingOptions)

func WithOperationNameFunc(opNameFunc func() string) TracingOption {
	return func(options *tracingOptions) {
		options.opNameFunc = opNameFunc
	}
}

func WithResultBody(resultBody bool) TracingOption {
	return func(options *tracingOptions) {
		options.resultBody = resultBody
	}
}

func WithMaxBodyLogSize(maxBodyLogSize int) TracingOption {
	return func(options *tracingOptions) {
		options.maxBodyLogSize = maxBodyLogSize
	}
}
