package registry

import (
	"context"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/naming"
	"github.com/molon/pkg/errors"
	"github.com/molon/pkg/plog"
	"golang.org/x/time/rate"
	gnaming "google.golang.org/grpc/naming"
)

const registerRetryRate = 1

type Register struct {
	doneC chan struct{}

	ctx    context.Context
	cancel context.CancelFunc
}

func registerSession(c *clientv3.Client, prefix string, addr string, ttl int) (*Session, error) {

	ss, err := NewSession(c, WithTTL(ttl), WithContext(c.Ctx()))
	if err != nil {
		return nil, err
	}

	gr := &naming.GRPCResolver{Client: c}
	if err = gr.Update(c.Ctx(), prefix, gnaming.Update{Op: gnaming.Add, Addr: addr}, clientv3.WithLease(ss.Lease())); err != nil {
		return nil, errors.WithStack(err)
	}

	plog.Infof("Registered \"%s/%s\" with %d-second lease", prefix, addr, ttl)
	return ss, nil
}

func NewRegister(c *clientv3.Client, prefix string, addr string, ttl int) *Register {
	doneC := make(chan struct{})
	ctx, cancel := context.WithCancel(c.Ctx())
	r := &Register{
		doneC:  doneC,
		ctx:    ctx,
		cancel: cancel,
	}

	go func() {
		defer close(doneC)

		rm := rate.NewLimiter(rate.Limit(registerRetryRate), registerRetryRate)
		for rm.Wait(ctx) == nil {
			ss, err := registerSession(c, prefix, addr, ttl)
			if err != nil {
				plog.Warnf("RegisterSession")
				continue
			}

			select {
			case <-ctx.Done():
				err := ss.Close()
				if err != nil {
					plog.Warn("Ctx done, session close")
				}
				return

			case <-ss.Done():
				plog.Warn("Session expired; possible network partition or server restart")
				plog.Warn("Creating a new session to rejoin")
				continue
			}
		}
	}()

	return r
}

func (r *Register) Done() <-chan struct{} { return r.doneC }

func (r *Register) Close() {
	r.cancel()
	<-r.doneC
}
