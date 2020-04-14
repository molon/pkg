package putil

import (
	"context"
	"time"

	"github.com/molon/pkg/errors"
)

func TimeUnixMilli(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func TimeNowUnixMilli(t time.Time) int64 {
	return TimeUnixMilli(time.Now())
}

func Sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	select {
	case <-ctx.Done():
		t.Stop()
		return errors.WithStack(ctx.Err())
	case <-t.C:
	}
	return nil
}
