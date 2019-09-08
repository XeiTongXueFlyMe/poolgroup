package group

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	ROLLBACK_MAXNUM     = 10000
	GO_F_TYPE_ERR       = "g.Go(f): f.(type) is FAILURE"
	FUNC_CALL_LOGIC_ERR = "err :calling Configure after calling g.Go()"
	ROLLBACK_ERR        = "rollback ERR: "
)

type Group struct {
	child []*Group

	isUsed     bool
	counter    uint64
	errs       []error
	m          sync.Mutex
	wg         sync.WaitGroup
	ctx        *context.Context
	cancel     context.CancelFunc
	isRollback bool
	rollback   chan func() error
}

func NewGroup() *Group {
	return &Group{rollback: make(chan func() error, ROLLBACK_MAXNUM)}
}
func (g *Group) ForkChild() *Group {
	g.m.Lock()
	defer g.m.Unlock()

	child := NewGroup()
	if g.ctx != nil {
		ctx, cancel := context.WithCancel(*g.ctx)
		child.ctx = &ctx
		child.cancel = cancel
	}
	g.child = append(g.child, child)

	return child
}

func (g *Group) WithContext(ctx context.Context) (c context.Context) {
	g.checkCallLogic()

	c, g.cancel = context.WithCancel(ctx)
	g.ctx = &c
	return
}

func (g *Group) WithTimeout(ctx context.Context, timeout time.Duration) (c context.Context) {
	g.checkCallLogic()

	c, g.cancel = context.WithTimeout(ctx, timeout)
	g.ctx = &c
	return
}

func (g *Group) DiscardedContext() {
	g.checkCallLogic()

	g.ctx = nil
	g.cancel = nil
	return
}

func (g *Group) Wait(isParentRollback ...interface{}) []error {
	var err []error

	for _, b := range isParentRollback {
		switch t := b.(type) {
		case bool:
			if g.ctx != nil {
				g.isRollback = (t || g.isRollback)
			}
		}
	}

	if g.wg.Wait(); g.cancel != nil {
		g.cancel()
	}

	if (len(g.errs) > 0) && (g.ctx != nil) {
		g.isRollback = true
	}

	for _, v := range g.child {
		e := v.Wait(g.isRollback)
		err = append(err, e...)
	}

	if e := g.callRollback(); len(e) > 0 {
		err = append(err, e...)
	}

	return append(err, g.errs...)
}

func (g *Group) Go(f interface{}, rollback ...interface{}) error {
	g.m.Lock()
	g.isUsed = true
	g.wg.Add(1)
	g.counter++
	g.m.Unlock()

	if g.ctx != nil {
		if _, ok := f.(func(ctx context.Context) error); !ok {
			return errors.New(GO_F_TYPE_ERR)
		}
		go g.fWithContext(f.(func(ctx context.Context) error), rollback...)
	} else {
		if _, ok := f.(func() error); !ok {
			return errors.New(GO_F_TYPE_ERR)
		}
		go g.f(f.(func() error))
	}
	return nil
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
	if (g.ctx != nil) && (g.cancel != nil) {
		g.cancel()
		g.isRollback = true
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
func (g *Group) fWithContext(f func(ctx context.Context) error, rollback ...interface{}) {
	defer g.wg.Done()
	defer func() {
		for _, v := range rollback {
			f, ok := v.(func() error)
			if ok {
				g.rollback <- f
			}
		}
	}()
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
		panic(FUNC_CALL_LOGIC_ERR)
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

//当并发线程某一个返回错误,或则panic时 执行回滚
//父亲组产生回滚,子组树全部产生回滚,不带上下文的节点及派生的子树不回滚
//子组产生回滚，其父不回滚
func (g *Group) callRollback() []error {
	var err []error
	if g.isRollback {
		for {
			select {
			case f := <-g.rollback:
				if e := f(); e != nil {
					err = append(err, errors.New(fmt.Sprintln(ROLLBACK_ERR, e)))
				}
				continue
			default:
			}
			break
		}
	}
	return err
}
