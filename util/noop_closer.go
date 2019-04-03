package util

type NoopCloser struct{}

func (NoopCloser) Close() error {
	return nil
}
