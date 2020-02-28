package putil

import (
	"sync"
)

type LoopTomb struct {
	c chan struct{}
	o sync.Once
	w sync.WaitGroup
}

func NewLoopTomb() *LoopTomb {
	return &LoopTomb{c: make(chan struct{})}
}

func (t *LoopTomb) stop()  { t.o.Do(func() { close(t.c) }) }
func (t *LoopTomb) Close() { t.stop(); t.w.Wait() }
func (t *LoopTomb) Wait()  { t.w.Wait() }

func (t *LoopTomb) Dying() <-chan struct{} { return t.c }

// 如果f返回false，则表示某Goroutine并未正常执行逻辑，需要关闭所有Goroutine
func (t *LoopTomb) Go(f func(<-chan struct{}) bool) {
	t.w.Add(1)

	go func() {
		defer t.w.Done()

		if f(t.c) == false {
			t.stop()
		}
	}()
}
