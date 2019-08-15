package poolgroup

import (
	"context"
	"sync"
	"time"
)

//TODO：做一个父亲群组
type ErrGroup struct {
	wg      sync.WaitGroup
	counter uint64
	ctx     *context.Context
	m       sync.Mutex
	errs    []error
	cancel  func()
}

func NewErrGroup() *ErrGroup {
	return &ErrGroup{}
}

func (g *ErrGroup) WithContext(ctx context.Context) (c context.Context) {
	c, g.cancel = context.WithCancel(ctx)
	g.ctx = &c
	return
}

func (g *ErrGroup) WithTimeout(ctx context.Context, microSecond time.Duration) (c context.Context) {
	c, g.cancel = context.WithTimeout(ctx, microSecond*time.Microsecond)
	g.ctx = &c
	return
}

func (g *ErrGroup) Wait() []error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return g.errs
}

func (g *ErrGroup) Go(f func() error) {
	g.m.Lock()
	g.wg.Add(1)
	g.counter++
	g.m.Unlock()

	if g.ctx != nil {
		g.fWithContext(f)
	} else {
		go g.f(f)
	}
}

func (g *ErrGroup) GetGoroutineNum() uint64 {
	g.m.Lock()
	defer g.m.Unlock()

	return g.counter
}

//only to WithContext()  WithTimeout()
func (g *ErrGroup) Close() {
	if g.ctx != nil {
		g.Go(func() error { return nil })
	}
	return
}

func (g *ErrGroup) f(f func() error) {
	defer g.wg.Done()

	if err := f(); err != nil {
		g.m.Lock()
		g.errs = append(g.errs, err)
		g.m.Unlock()

		if g.cancel != nil {
			g.cancel()
		}
	}
}
func (g *ErrGroup) fWithContext(f func() error) {
	defer g.wg.Done()

	select {
	case <-(*g.ctx).Done():
		return
	default:
	}

	if err := f(); err != nil {
		g.m.Lock()
		g.errs = append(g.errs, err)
		g.m.Unlock()

		if g.cancel != nil {
			g.cancel()
		}
	}
}
