package clientstore

import (
	"context"
	"encoding/json"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/molon/pkg/errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/naming"
	"google.golang.org/grpc/status"

	"github.com/coreos/etcd/mvcc/mvccpb"
)

var ErrWatcherClosed = status.Errorf(codes.Unavailable, "naming: watch closed")

type Update struct {
	Op       naming.Operation
	Addr     string
	Target   string
	Metadata interface{}
}

func NewGRPCWatcher(c *etcd.Client, targetPrefix string) *GRPCWatcher {
	ctx, cancel := context.WithCancel(context.Background())
	w := &GRPCWatcher{c: c, targetPrefix: targetPrefix, ctx: ctx, cancel: cancel}
	return w
}

type GRPCWatcher struct {
	c            *etcd.Client
	targetPrefix string
	ctx          context.Context
	cancel       context.CancelFunc
	wch          etcd.WatchChan
	err          error
}

func (gw *GRPCWatcher) Next() ([]*Update, error) {
	if gw.wch == nil {
		return gw.firstNext()
	}
	if gw.err != nil {
		return nil, gw.err
	}

	wr, ok := <-gw.wch
	if !ok {
		gw.err = errors.WithStack(ErrWatcherClosed)
		return nil, gw.err
	}
	if gw.err = wr.Err(); gw.err != nil {
		return nil, gw.err
	}

	updates := make([]*Update, 0, len(wr.Events))
	for _, e := range wr.Events {
		var jupdate *Update
		var err error
		switch e.Type {
		case etcd.EventTypePut:
			jupdate, err = unmarshalToUpdate(e.Kv)
			jupdate.Op = naming.Add
		case etcd.EventTypeDelete:
			jupdate, err = unmarshalToUpdate(e.PrevKv)
			jupdate.Op = naming.Delete
		}
		if err == nil {
			updates = append(updates, jupdate)
		}
	}
	return updates, nil
}

func (gw *GRPCWatcher) firstNext() ([]*Update, error) {
	resp, err := gw.c.Get(gw.ctx, gw.targetPrefix, etcd.WithPrefix(), etcd.WithSerializable())
	if gw.err = err; err != nil {
		return nil, errors.WithStack(err)
	}

	updates := make([]*Update, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		jupdate, err := unmarshalToUpdate(kv)
		if err != nil {
			continue
		}
		jupdate.Op = naming.Add

		updates = append(updates, jupdate)
	}

	opts := []etcd.OpOption{etcd.WithRev(resp.Header.Revision + 1), etcd.WithPrefix(), etcd.WithPrevKV()}
	gw.wch = gw.c.Watch(gw.ctx, gw.targetPrefix, opts...)
	return updates, nil
}

func (gw *GRPCWatcher) Close() { gw.cancel() }

func unmarshalToUpdate(kv *mvccpb.KeyValue) (*Update, error) {
	var jupdate Update
	if err := json.Unmarshal(kv.Value, &jupdate); err != nil {
		return nil, errors.WithStack(err)
	}

	// key会是 msg://boat/[::]:51841 这种，target名称是 msg://boat
	key := string(kv.Key)
	jupdate.Target = key[:len(key)-len(jupdate.Addr)-1]
	return &jupdate, nil
}
