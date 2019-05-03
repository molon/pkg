package util

import "sync"

type LoopLife struct {
	c chan struct{}
	o sync.Once
	w sync.WaitGroup
}

// LoopTomb里的任务只要有一个结束了，所有的都会结束
// LoopLife里的任务并不会，且额外提供了Wait()方法便于阻塞等待所有任务的正常执行结束
func NewLoopLife() *LoopLife {
	return &LoopLife{c: make(chan struct{})}
}

func (t *LoopLife) Dying() <-chan struct{} { return t.c }
func (t *LoopLife) Close() {
	t.o.Do(func() {
		close(t.c)
	})
	t.Wait()
}

func (t *LoopLife) Wait() { t.w.Wait() }
func (t *LoopLife) Go(f func(<-chan struct{})) {
	t.w.Add(1)

	go func() {
		defer t.w.Done()

		f(t.c)
	}()
}
