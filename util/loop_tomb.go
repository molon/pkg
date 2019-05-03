package util

import "sync"

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

func (t *LoopTomb) Dying() <-chan struct{} { return t.c }
func (t *LoopTomb) Go(f func(<-chan struct{})) {
	t.w.Add(1)

	go func() {
		defer t.stop()
		defer t.w.Done()

		f(t.c)
	}()
}
