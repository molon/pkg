package putil

import (
	"context"
	"time"

	"github.com/molon/pkg/errors"
)

func TimeUnixMilli(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func TimeUnixMilliNow() int64 {
	return TimeUnixMilli(time.Now())
}

func TimeFromUnixMilli(u int64) time.Time {
	return time.Unix(0, u*int64(time.Millisecond))
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
