package clientstore

import (
	"context"
	"io"
	"sync"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"

	etcd "github.com/coreos/etcd/clientv3"
	etcdnaming "github.com/coreos/etcd/clientv3/naming"
)

type DialFunc func(target string, opts ...grpc.DialOption) (interface{}, io.Closer, error)

type Store struct {
	mu sync.RWMutex

	logger       *logrus.Entry
	etcdCli      *etcd.Client
	targetPrefix string
	dial         DialFunc

	// 服务发现
	w *GRPCWatcher

	// 当前存在的客户端
	targetToClient map[string]*client

	// 当前发现的target和地址记录
	targetToAddrs map[string][]grpc.Address
}

func NewStore(
	logger *logrus.Logger,
	etcdCli *etcd.Client,
	targetPrefix string,
	dial DialFunc,
) *Store {
	ll := logger.WithFields(logrus.Fields{
		"pkg": "clientstore",
		"mod": "store",
	})

	return &Store{
		logger:         ll,
		etcdCli:        etcdCli,
		targetPrefix:   targetPrefix,
		dial:           dial,
		w:              NewGRPCWatcher(etcdCli, targetPrefix),
		targetToAddrs:  make(map[string][]grpc.Address),
		targetToClient: make(map[string]*client),
	}
}

func (cs *Store) Start() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	go func() {
		for {
			// 在watcher关闭之后会触发err
			if err := cs.watchAddrUpdates(); err != nil {
				if err != ErrWatcherClosed {
					cs.logger.WithError(err).Warnf("watchAddrUpdates")
				}
				return
			}
		}
	}()
	return nil
}

func (cs *Store) Stop() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// 关闭watcher
	cs.w.Close()

	// 关闭连接
	for _, client := range cs.targetToClient {
		client.close()
	}
}

func (cs *Store) Get(target string) (interface{}, bool) {
	cs.mu.RLock()
	client, ok := cs.targetToClient[target]
	cs.mu.RUnlock()

	if ok {
		return client.cli(), ok
	}

	return nil, false
}

func (cs *Store) watchAddrUpdates() error {
	updates, err := cs.w.Next()
	if err != nil {
		return err
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	for _, update := range updates {
		if len(update.Target) < 1 {
			continue
		}

		target := update.Target

		address := grpc.Address{
			Addr:     update.Addr,
			Metadata: update.Metadata,
		}

		switch update.Op {
		case naming.Add:
			var exist bool
			for _, addr := range cs.targetToAddrs[target] {
				if addr == address {
					exist = true
					cs.logger.Infoln("The name resolver wanted to add an existing address: ", addr)
					break
				}
			}
			if exist {
				continue
			}

			cs.targetToAddrs[target] = append(cs.targetToAddrs[target], address)
		case naming.Delete:
			addrs, ok := cs.targetToAddrs[target]
			if ok {
				for i, addr := range addrs {
					if addr == address {
						copy(addrs[i:], addrs[i+1:])
						addrs = addrs[:len(addrs)-1]
						break
					}
				}
				if len(addrs) <= 0 {
					delete(cs.targetToAddrs, target)
				} else {
					cs.targetToAddrs[target] = addrs
				}
			}
		default:
			cs.logger.Errorln("Unknown update.Op ", update.Op)
		}
	}

	// targetToClient 有 targetToAddrs 无，则减少
	for target, client := range cs.targetToClient {
		_, ok := cs.targetToAddrs[target]
		if !ok {
			delete(cs.targetToClient, target)
			client.close()
		}
	}

	// targetToAddrs 有 targetToClient 无，则增加
	for target, _ := range cs.targetToAddrs {
		_, ok := cs.targetToClient[target]
		if !ok {
			r := &etcdnaming.GRPCResolver{Client: cs.etcdCli}
			opt := grpc.WithBalancer(grpc.RoundRobin(r))

			cs.targetToClient[target] = newClient(context.Background(), cs.logger, target,
				func() (interface{}, io.Closer, error) {
					return cs.dial(target, opt)
				},
			)
		}
	}

	return nil
}
