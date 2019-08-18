package group

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

//TODO：支持回滚
type Group struct {
	child   []*Group
	isUsed  bool
	wg      sync.WaitGroup
	counter uint64
	ctx     *context.Context
	m       sync.Mutex
	errs    []error
	cancel  context.CancelFunc
}

func NewGroup() *Group {
	return &Group{}
}
func (g *Group) ForkChild() *Group {
	g.m.Lock()
	defer g.m.Unlock()

	child := &Group{ctx: g.ctx}
	g.child = append(g.child, child)

	return child
}

func (g *Group) WithContext(ctx context.Context) (c context.Context) {
	g.checkCallLogic()

	c, g.cancel = context.WithCancel(ctx)
	g.ctx = &c
	return
}

func (g *Group) WithTimeout(ctx context.Context, microSecond time.Duration) (c context.Context) {
	g.checkCallLogic()

	c, g.cancel = context.WithTimeout(ctx, microSecond*time.Microsecond)
	g.ctx = &c
	return
}

func (g *Group) DiscardedContext() {
	g.checkCallLogic()

	g.ctx = nil
	g.cancel = nil
	return
}

func (g *Group) Wait() []error {
	for _, v := range g.child {
		v.Wait()
	}

	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return g.errs
}

func (g *Group) Go(f interface{}) {
	g.m.Lock()
	g.isUsed = true
	g.wg.Add(1)
	g.counter++
	g.m.Unlock()

	if g.ctx != nil {
		go g.fWithContext(f.(func(ctx context.Context) error))
	} else {
		go g.f(f.(func() error))
	}
}

func (g *Group) GetGoroutineNum() uint64 {
	var n uint64
	for _, v := range g.child {
		n = n + v.GetGoroutineNum()
	}

	g.m.Lock()
	defer g.m.Unlock()

	return n + g.counter
}

func (g *Group) GetErrs() []error {
	var e []error
	for _, v := range g.child {
		e = append(e, v.GetErrs()...)
	}

	g.m.Lock()
	defer g.m.Unlock()

	return append(e, g.errs...)
}

//only to use WithContext()  WithTimeout()
func (g *Group) Close() {
	if g.ctx != nil {
		g.cancel()
	}
	return
}

//func (g *Group) f(f func() error) {
func (g *Group) f(f func() error) {
	defer g.wg.Done()
	defer func() {
		if e := recover(); e != nil {
			g.collectErrs(errors.New(fmt.Sprint(e)))
		}
	}()

	if e := f(); e != nil {
		g.collectErrs(e)
	}
}
func (g *Group) fWithContext(f func(ctx context.Context) error) {
	defer g.wg.Done()
	defer func() {
		if e := recover(); e != nil {
			g.collectErrs(errors.New(fmt.Sprint(e)))
		}
	}()

	if e := f(*g.ctx); e != nil {
		g.collectErrs(e)
	}
}

func (g *Group) checkCallLogic() {
	g.m.Lock()
	defer g.m.Unlock()

	//并发产生后，不能再配置Group
	if g.isUsed {
		panic("err :calling Configure after calling g.Go()")
	}
}

func (g *Group) collectErrs(err error) {
	g.m.Lock()
	defer g.m.Unlock()
	g.errs = append(g.errs, err)

	if g.cancel != nil {
		g.cancel()
	}
}
