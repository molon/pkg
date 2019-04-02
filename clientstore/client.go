package clientstore

import (
	"context"
	"io"
	"sync"

	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

const dialRetryRate = 1

type client struct {
	ctx    context.Context
	cancel context.CancelFunc
	doneC  chan struct{}

	mu sync.RWMutex
	cc interface{}
}

func newClient(ctx context.Context, logger *logrus.Entry, target string, dial func() (interface{}, io.Closer, error)) *client {
	ll := logger.WithFields(logrus.Fields{
		"mod": "client",
	})

	doneC := make(chan struct{})
	ctx, cancel := context.WithCancel(ctx)
	c := &client{
		ctx:    ctx,
		cancel: cancel,
		doneC:  doneC,
	}

	go func() {
		defer close(doneC)

		rm := rate.NewLimiter(rate.Limit(dialRetryRate), dialRetryRate)
		for rm.Wait(ctx) == nil {
			cc, closer, err := dial()
			if err != nil {
				ll.WithError(err).Warningf("Dial")
				continue
			}

			c.mu.Lock()
			c.cc = cc
			c.mu.Unlock()

			ll.Infof("Dial %q succeed", target)

			select {
			case <-ctx.Done():
				err := closer.Close()
				if err != nil {
					ll.WithError(err).Warningf("Ctx done, Close %q failed", target)
				} else {
					ll.Infof("Close %q succeed", target)
				}
				return
			}
		}
	}()

	return c
}

func (c *client) cli() interface{} {
	// TODO: 这里是否要做为nil时阻塞呢
	c.mu.RLock()
	cli := c.cc
	c.mu.RUnlock()
	return cli
}

func (c *client) done() <-chan struct{} { return c.doneC }

func (c *client) close() {
	c.cancel()
	<-c.doneC
}
