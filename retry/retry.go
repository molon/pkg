package retry

import (
	"context"
	"fmt"
	stdlog "log"
	"math"
	"strings"
	"time"

	"github.com/molon/pkg/errors"
)

type Status uint8

const (
	StatusUnknown Status = iota
	StatusWaitBegin
	StatusWaitEnd
	StatusBegin
	StatusEnd
)

var ErrorContinue = errors.New("ErrorContinue")

func NewLog(printf func(format string, v ...interface{})) func(context.Context, string, int, Status, error) {
	return func(ctx context.Context, name string, idx int, s Status, err error) {
		if printf == nil {
			return
		}

		if name == "" {
			return
		}

		var tips string
		switch s {
		case StatusWaitBegin:
			// tips = fmt.Sprintf("%s Wait...", name)
		case StatusWaitEnd:
			if err != nil {
				tips = fmt.Sprintf("%s Wait Failed: %v", name, err)
			}
		case StatusBegin:
			tips = fmt.Sprintf("%s ...", name)
		case StatusEnd:
			if err != nil {
				tips = fmt.Sprintf("%s Failed: %v", name, err)
			} else {
				tips = fmt.Sprintf("%s Successfully !", name)
			}
		}

		if tips == "" {
			return
		}

		if idx > 0 {
			tips = fmt.Sprintf("#%d %s", idx, tips)
		}
		printf("%s", tips)
	}
}

func NewLevelLog(printf func(good bool, format string, v ...interface{})) func(context.Context, string, int, Status, error) {
	return func(ctx context.Context, name string, idx int, s Status, err error) {
		if printf == nil {
			return
		}

		if name == "" {
			return
		}

		var tips string
		good := true
		switch s {
		case StatusWaitBegin:
			// tips = fmt.Sprintf("%s Wait...", name)
		case StatusWaitEnd:
			if err != nil {
				tips = fmt.Sprintf("%s Wait Failed: %v", name, err)
				good = false
			}
		case StatusBegin:
			tips = fmt.Sprintf("%s ...", name)
		case StatusEnd:
			if err != nil {
				tips = fmt.Sprintf("%s Failed: %v", name, err)
				good = false
			} else {
				tips = fmt.Sprintf("%s Successfully !", name)
			}
		}

		if tips == "" {
			return
		}

		if idx > 0 {
			tips = fmt.Sprintf("#%d %s", idx, tips)
		}
		printf(good, "%s", tips)
	}
}

type Retry struct {
	n       int
	delay   time.Duration
	timeout time.Duration

	log    func(ctx context.Context, name string, idx int, s Status, err error)
	wait   func(ctx context.Context, name string, idx int) error
	fix    func(ctx context.Context, name string, idx int, err error) error
	deferr func(ctx context.Context, name string, idx int, err error) error
}

func (r *Retry) Clone() *Retry {
	return &Retry{
		n:       r.n,
		delay:   r.delay,
		timeout: r.timeout,
		log:     r.log,
		wait:    r.wait,
		fix:     r.fix,
		deferr:  r.deferr,
	}
}

func New(opts ...Option) *Retry {
	r := &Retry{
		n:       math.MaxInt32,
		delay:   0,
		timeout: 0,

		log: NewLog(stdlog.Printf),
	}

	for _, o := range opts {
		o(r)
	}

	return r
}

func (r *Retry) Do(ctx context.Context, name string, f func(ctx context.Context, idx int) error, opts ...Option) error {
	// 每次都clone一个新的出来
	r = r.Clone()
	for _, o := range opts {
		o(r)
	}

	var finalErr error
	one := func(idx int) (xerr error) {
		defer func() {
			if r.deferr != nil {
				xerr = r.deferr(ctx, name, idx, xerr)
			}
		}()

		var err error
		defer func() {
			finalErr = err
		}()

		if r.wait != nil {
			r.log(ctx, name, idx, StatusWaitBegin, nil)
			err = r.wait(ctx, name, idx)
			r.log(ctx, name, idx, StatusWaitEnd, err)
		}

		if err == nil {
			r.log(ctx, name, idx, StatusBegin, nil)
			var tctx context.Context
			var cancel context.CancelFunc
			if r.timeout > 0 {
				tctx, cancel = context.WithTimeout(ctx, r.timeout)
			} else {
				tctx, cancel = context.WithCancel(ctx)
			}
			defer cancel()
			err = f(tctx, idx)
			r.log(ctx, name, idx, StatusEnd, err)
		}

		if err == nil {
			return nil
		}

		// 如果ctx结束了，则直接无需重试了
		if ctx.Err() != nil {
			return err // 考虑了一下还是返回原错误会比较合适一些
		}
		/*
			如果用到限流器，可能会遇到 fmt.Errorf("rate: Wait(n=%d) would exceed context deadline", n) 错误，这是因为ctx的deadline剩余时间不足以拿到token了，所以就直接failed了，但是因为我们这是重试管理器，默认无delay重试，所以就会在限流器的拿token时间内导致死循环无限，直到ctx.Deadline触发。
			但是我们又不能完全确定说限流器触发此错误是由于当前的这个ctx，(例如有可能是f内部另外单独开辟的ctx limiter导致的)
			所以暂且只能针对此情况主动delay一小段时间，防止CPU资源过度使用
		*/
		if strings.HasPrefix(errors.Cause(err).Error(), "rate: Wait(n=") { // 只判断前缀是因为rate package里还有类似的其他错误也最好delay一下
			t := time.NewTimer(100 * time.Millisecond)
			select {
			case <-ctx.Done():
				t.Stop()
				return errors.WithStack(ctx.Err())
			case <-t.C:
			}
		}

		if idx < r.n-1 { // 非最后一个再进行如下逻辑
			// 可能需要做修正，例如更换代理之类的
			// 如果修正失败，就直接返回了
			if r.fix != nil {
				ferr := r.fix(ctx, name, idx, err)
				if ferr != nil {
					return ferr
				}
			}

			// 非最后一个就等待延时重试
			if r.delay > 0 {
				t := time.NewTimer(r.delay)
				select {
				case <-ctx.Done():
					t.Stop()
					return errors.WithStack(ctx.Err())
				case <-t.C:
				}
			}
		}
		return ErrorContinue
	}

	for idx := 0; idx < r.n; idx++ {
		err := one(idx)
		if err == ErrorContinue {
			continue
		}
		return err
	}
	return finalErr // 如果上述N次执行完毕仍未完成，则返回最后一个错误
}

func Do(ctx context.Context, name string, f func(ctx context.Context, idx int) error, opts ...Option) error {
	return New(opts...).Do(ctx, name, f)
}

type Option func(*Retry)

func WithN(n int) Option {
	return func(r *Retry) {
		r.n = n
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(r *Retry) {
		r.timeout = timeout
	}
}

func WithDelay(delay time.Duration) Option {
	return func(r *Retry) {
		r.delay = delay
	}
}

func WithLog(log func(ctx context.Context, name string, idx int, s Status, err error)) Option {
	return func(r *Retry) {
		r.log = log
	}
}

func WithWait(wait func(ctx context.Context, name string, idx int) error) Option {
	return func(r *Retry) {
		r.wait = wait
	}
}

func WithFix(fix func(ctx context.Context, name string, idx int, err error) error) Option {
	return func(r *Retry) {
		r.fix = fix
	}
}

func WithDefer(deferr func(ctx context.Context, name string, idx int, err error) error) Option {
	return func(r *Retry) {
		r.deferr = deferr
	}
}
