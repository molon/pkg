package putil

import (
	"context"
	"time"

	"github.com/molon/pkg/errors"
)

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
