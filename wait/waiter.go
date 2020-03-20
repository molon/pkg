package wait

import (
	"context"
	"sync"
	"time"

	"github.com/molon/pkg/errors"
)

var ErrorWaiterClosed = errors.New("ErrorWaiterClosed")
var ErrorWaiterContinue = errors.New("ErrorWaiterContinue")

type Waiter struct {
	Lock sync.RWMutex

	closed bool
	cond   sync.Cond
}

func NewWaiter() *Waiter {
	w := &Waiter{}
	w.cond.L = &w.Lock
	return w
}

func (w *Waiter) Close(f func() error) error {
	w.Lock.Lock()
	defer w.Lock.Unlock()

	if f != nil {
		if err := f(); err != nil {
			return err
		}
	}

	w.closed = true
	w.cond.Broadcast()
	return nil
}

func (w *Waiter) Broadcast(f func() error) error {
	w.Lock.Lock()
	defer w.Lock.Unlock()
	if f != nil {
		if err := f(); err != nil {
			return err
		}
	}
	w.cond.Broadcast()
	return nil
}

func (w *Waiter) Wait(ctx context.Context, query func(context.Context) error) error {
	hasCtxMonitor := false
	ctxMonitorCloseC := make(chan struct{})
	ctxMonitorDoneC := make(chan struct{})
	defer func() {
		if hasCtxMonitor {
			close(ctxMonitorCloseC)
			<-ctxMonitorDoneC
		}
	}()

	w.Lock.Lock()
	defer w.Lock.Unlock()

	idx := 0
	for {
		if ctx.Err() != nil {
			return errors.WithStack(ctx.Err())
		}
		if w.closed {
			return errors.WithStack(ErrorWaiterClosed)
		}

		err := query(ctx) // 找到了或者出错了就返回
		if err == nil || errors.Cause(err) != ErrorWaiterContinue {
			return err
		}

		if idx == 0 {
			hasCtxMonitor = true
			go func() {
				defer close(ctxMonitorDoneC)
				for {
					select {
					case <-ctx.Done():
						w.Lock.Lock()
						w.cond.Broadcast()
						w.Lock.Unlock()
						// 100ms检测一次是为了防止这个Broadcast在下面的Wait之前执行了，几乎不可能
						timer := time.NewTimer(100 * time.Millisecond)
						select {
						case <-ctxMonitorCloseC:
							timer.Stop()
							return
						case <-timer.C:
						}
					case <-ctxMonitorCloseC:
						return
					}
				}
			}()
		}
		w.cond.Wait()
		idx++
	}
}
