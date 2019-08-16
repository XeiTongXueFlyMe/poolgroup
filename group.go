package poolgroup

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

//TODO：做一个父亲群组
//TODO：支持回滚
//TODO:阶段painc()
type ErrGroup struct {
	wg      sync.WaitGroup
	counter uint64
	ctx     *context.Context
	m       sync.Mutex
	errs    []error
	cancel  context.CancelFunc
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
		go g.fWithContext(f)
	} else {
		go g.f(f)
	}
}

func (g *ErrGroup) GetGoroutineNum() uint64 {
	g.m.Lock()
	defer g.m.Unlock()

	return g.counter
}

//only to use WithContext()  WithTimeout()
func (g *ErrGroup) Close() {
	if g.ctx != nil {
		g.Go(func() error { return nil })
	}
	return
}

func (g *ErrGroup) f(f func() error) {
	defer g.wg.Done()
	defer func() {
		if err := recover(); err != nil {
			g.collectErrs(errors.New(fmt.Sprint(err)))
		}
	}()

	if err := f(); err != nil {
		g.collectErrs(err)
	}
}
func (g *ErrGroup) fWithContext(f func() error) {
	defer g.wg.Done()
	defer func() {
		if err := recover(); err != nil {
			g.collectErrs(errors.New(fmt.Sprint(err)))
		}
	}()

	if err := f(); err != nil {
		g.collectErrs(err)
	}
}

func (g *ErrGroup) collectErrs(err error) {
	g.m.Lock()
	defer g.m.Unlock()
	g.errs = append(g.errs, err)

	if g.cancel != nil {
		g.cancel()
	}
}
