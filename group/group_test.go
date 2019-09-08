package group

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestGroup(t *testing.T) {
	counts, m := 0, sync.Mutex{}

	g := NewGroup()
	assert.NoError(t, g.Go(func() error {
		time.Sleep(100 * time.Millisecond)
		m.Lock()
		defer m.Unlock()
		counts++
		return errors.New("err")
	}))
	assert.NoError(t, g.Go(func() error {
		time.Sleep(200 * time.Millisecond)
		m.Lock()
		defer m.Unlock()
		counts++
		return nil
	}))
	assert.NoError(t, g.Go(func() error {
		time.Sleep(300 * time.Millisecond)
		m.Lock()
		defer m.Unlock()
		counts++
		return errors.New("err")
	}))
	g.Wait()
	for _, err := range g.errs {
		assert.Error(t, err)
	}
	assert.EqualValues(t, 3, g.GetGoroutineNum())
	assert.EqualValues(t, 3, counts)
}

func TestGroupPanic(t *testing.T) {
	counts, m := 0, sync.Mutex{}

	g := NewGroup()
	assert.NoError(t, g.Go(func() error {
		time.Sleep(100 * time.Millisecond)
		m.Lock()
		defer m.Unlock()
		counts++
		return errors.New("err")
	}))
	assert.NoError(t, g.Go(func() error {
		time.Sleep(200 * time.Millisecond)
		m.Lock()
		defer m.Unlock()
		counts++
		panic("panic")
	}))
	assert.NoError(t, g.Go(func() error {
		time.Sleep(300 * time.Millisecond)
		m.Lock()
		defer m.Unlock()
		counts++
		return errors.New("err")
	}))
	g.Wait()
	for _, err := range g.errs {
		assert.Error(t, err)
	}
	assert.EqualValues(t, 3, g.GetGoroutineNum())
	assert.EqualValues(t, 3, counts)
}

type class struct {
	a int
	b int
	c int
	m sync.Mutex
}

func (this *class) funcA() error {
	time.Sleep(100 * time.Millisecond)
	this.m.Lock()
	defer this.m.Unlock()
	this.a++

	return nil
}
func (this *class) funcResetA() error {
	time.Sleep(200 * time.Millisecond)
	this.m.Lock()
	defer this.m.Unlock()
	this.a = 0

	return nil
}
func (this *class) funcAA() error {
	time.Sleep(200 * time.Millisecond)
	this.m.Lock()
	defer this.m.Unlock()
	this.a++

	return nil
}
func (this *class) funcB() error {
	time.Sleep(100 * time.Millisecond)
	this.m.Lock()
	defer this.m.Unlock()
	this.b++

	return nil
}
func (this *class) funcResetB() error {
	time.Sleep(200 * time.Millisecond)
	this.m.Lock()
	defer this.m.Unlock()
	this.b = 0

	return nil
}
func (this *class) funcC() error {
	time.Sleep(100 * time.Millisecond)
	this.m.Lock()
	defer this.m.Unlock()
	this.c++

	return nil
}
func (this *class) funcResetC() error {
	time.Sleep(200 * time.Millisecond)
	this.m.Lock()
	defer this.m.Unlock()
	this.c = 0
	return nil
}
func (this *class) funcCtxA(ctx context.Context) error {

	for {
		time.Sleep(200 * time.Millisecond)
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		this.m.Lock()
		this.a++
		this.m.Unlock()
	}
}
func (this *class) funcCtxB(ctx context.Context) error {

	for {
		time.Sleep(200 * time.Millisecond)
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		this.m.Lock()
		this.b++
		this.m.Unlock()
	}
}
func (this *class) funcCtxC(ctx context.Context) error {
	for {
		time.Sleep(200 * time.Millisecond)
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		this.m.Lock()
		this.c++
		this.m.Unlock()
	}
}
func (this *class) funcTimeOut(ctx context.Context) error {

	time.Sleep(300 * time.Millisecond)
	return errors.New("time out")
}
func (this *class) funcComplete(ctx context.Context, cancel context.CancelFunc) error {

	time.Sleep(300 * time.Millisecond)
	cancel()
	return nil
}

func TestGroupWait(t *testing.T) {
	f := class{}

	g := NewGroup()
	assert.NoError(t, g.Go(f.funcA))
	assert.NoError(t, g.Go(f.funcB))
	assert.NoError(t, g.Go(f.funcC))

	A := g.ForkChild()
	assert.NoError(t, A.Go(f.funcB))
	assert.NoError(t, A.Go(f.funcC))

	B := g.ForkChild()
	B.WithContext(context.TODO())
	assert.NoError(t, B.Go(f.funcCtxA))
	assert.NoError(t, B.Go(f.funcCtxC))
	assert.NoError(t, B.Go(f.funcCtxC))
	assert.NoError(t, B.Go(f.funcTimeOut))

	a := B.ForkChild()
	assert.NoError(t, a.Go(f.funcCtxA))
	assert.NoError(t, a.Go(f.funcCtxC))
	assert.NoError(t, a.Go(f.funcCtxC))
	assert.NoError(t, a.Go(f.funcTimeOut))

	b := A.ForkChild()
	assert.NoError(t, b.Go(f.funcB))
	assert.NoError(t, b.Go(f.funcC))

	g.Wait()
	assert.EqualValues(t, 3, f.a)
	assert.EqualValues(t, 3, f.b)
	assert.EqualValues(t, 7, f.c)
}

func TestGroupDiscardedContext(t *testing.T) {
	f := class{}

	g := NewGroup()
	g.WithContext(context.Background())
	assert.NoError(t, g.Go(f.funcCtxA))
	assert.NoError(t, g.Go(f.funcTimeOut))

	A := g.ForkChild()
	A.DiscardedContext()
	assert.NoError(t, A.Go(f.funcA))

	g.Wait()

	assert.EqualValues(t, 3, g.GetGoroutineNum())
	assert.EqualValues(t, 2, f.a)
}

func TestGroupTimeout(t *testing.T) {
	f := class{}

	g := NewGroup()
	g.WithTimeout(context.Background(), 100*time.Millisecond)
	assert.NoError(t, g.Go(f.funcCtxA))

	A := g.ForkChild()
	A.DiscardedContext()
	assert.NoError(t, A.Go(f.funcA))

	g.Wait()

	assert.EqualValues(t, 2, g.GetGoroutineNum())
	assert.EqualValues(t, 1, f.a)
}

func TestGroupClose(t *testing.T) {
	f := class{}

	g := NewGroup()
	g.WithContext(context.TODO())
	assert.NoError(t, g.Go(f.funcCtxA))

	A := g.ForkChild()
	assert.NoError(t, A.Go(f.funcCtxA))
	assert.NoError(t, A.Go(f.funcCtxB))

	a := A.ForkChild()
	a.DiscardedContext()
	assert.NoError(t, a.Go(f.funcAA))

	g.Close()
	g.Wait()

	assert.EqualValues(t, 4, g.GetGoroutineNum())
	assert.EqualValues(t, 1, f.a)
}

func TestGroupGoroutineNum(t *testing.T) {
	f := class{}

	g := NewGroup()
	assert.NoError(t, g.Go(f.funcA))
	assert.NoError(t, g.Go(f.funcB))
	assert.NoError(t, g.Go(f.funcC))

	A := g.ForkChild()
	assert.NoError(t, A.Go(f.funcB))
	assert.NoError(t, A.Go(f.funcC))

	B := g.ForkChild()
	B.WithContext(context.TODO())
	assert.NoError(t, B.Go(f.funcCtxA))
	assert.NoError(t, B.Go(f.funcCtxC))
	assert.NoError(t, B.Go(f.funcCtxC))
	assert.NoError(t, B.Go(f.funcTimeOut))

	a := B.ForkChild()
	assert.NoError(t, a.Go(f.funcCtxA))
	assert.NoError(t, a.Go(f.funcCtxC))
	assert.NoError(t, a.Go(f.funcCtxC))
	assert.NoError(t, a.Go(f.funcTimeOut))

	b := A.ForkChild()
	assert.NoError(t, b.Go(f.funcB))
	assert.NoError(t, b.Go(f.funcC))

	g.Wait()
	assert.EqualValues(t, 15, g.GetGoroutineNum())
	assert.EqualValues(t, 4, A.GetGoroutineNum())
	assert.EqualValues(t, 8, B.GetGoroutineNum())
	assert.EqualValues(t, 4, a.GetGoroutineNum())
	assert.EqualValues(t, 2, b.GetGoroutineNum())
}

func TestGroupGetErrs(t *testing.T) {
	err := []error{errors.New("time out"), errors.New("time out")}
	f := class{}

	g := NewGroup()
	assert.NoError(t, g.Go(f.funcA))
	assert.NoError(t, g.Go(f.funcB))
	assert.NoError(t, g.Go(f.funcC))

	A := g.ForkChild()
	assert.NoError(t, A.Go(f.funcB))
	assert.NoError(t, A.Go(f.funcC))

	B := g.ForkChild()
	B.WithContext(context.TODO())
	assert.NoError(t, B.Go(f.funcCtxA))
	assert.NoError(t, B.Go(f.funcCtxC))
	assert.NoError(t, B.Go(f.funcCtxC))
	assert.NoError(t, B.Go(f.funcTimeOut))

	a := B.ForkChild()
	assert.NoError(t, a.Go(f.funcCtxA))
	assert.NoError(t, a.Go(f.funcCtxC))
	assert.NoError(t, a.Go(f.funcCtxC))
	assert.NoError(t, a.Go(f.funcTimeOut))

	b := A.ForkChild()
	assert.NoError(t, b.Go(f.funcB))
	assert.NoError(t, b.Go(f.funcC))

	g.Wait()
	assert.EqualValues(t, g.GetErrs(), err)
}

func TestGroupRollback_0(t *testing.T) {
	f := class{}

	g := NewGroup()
	g.WithContext(context.TODO())
	assert.NoError(t, g.Go(f.funcCtxA))
	assert.NoError(t, g.Go(f.funcCtxA, f.funcResetA))
	assert.NoError(t, g.Go(f.funcCtxB))
	assert.NoError(t, g.Go(f.funcCtxC, f.funcResetC))
	assert.NoError(t, g.Go(f.funcTimeOut))

	g.Wait()
	assert.EqualValues(t, 0, f.a)
	assert.EqualValues(t, 1, f.b)
	assert.EqualValues(t, 0, f.c)
}
func TestGroupRollback_1(t *testing.T) {
	f := class{}

	g := NewGroup()
	g.WithContext(context.TODO())
	assert.NoError(t, g.Go(f.funcCtxA))
	assert.NoError(t, g.Go(f.funcCtxA))
	assert.NoError(t, g.Go(f.funcCtxB))
	assert.NoError(t, g.Go(f.funcCtxC))

	a := g.ForkChild()
	assert.NoError(t, a.Go(f.funcCtxA))
	assert.NoError(t, a.Go(f.funcCtxA, f.funcResetA))
	assert.NoError(t, a.Go(f.funcCtxB))
	assert.NoError(t, a.Go(f.funcCtxC, f.funcResetC))

	a.Close()
	time.AfterFunc(time.Millisecond*300, g.Close)
	g.Wait()

	assert.EqualValues(t, 0, f.a)
	assert.EqualValues(t, 1, f.b)
	assert.EqualValues(t, 0, f.c)
}
func TestGroupRollback_2(t *testing.T) {
	f := class{}

	g := NewGroup()
	g.WithContext(context.TODO())
	assert.NoError(t, g.Go(f.funcCtxA))
	assert.NoError(t, g.Go(f.funcCtxA))
	assert.NoError(t, g.Go(f.funcCtxB))
	assert.NoError(t, g.Go(f.funcCtxC))
	assert.NoError(t, g.Go(f.funcTimeOut))

	A := g.ForkChild()
	assert.NoError(t, A.Go(f.funcCtxA))
	assert.NoError(t, A.Go(f.funcCtxA))
	assert.NoError(t, A.Go(f.funcCtxB))
	assert.NoError(t, A.Go(f.funcCtxC))

	a := A.ForkChild()
	assert.NoError(t, a.Go(f.funcCtxA))
	assert.NoError(t, a.Go(f.funcCtxA, f.funcResetA))
	assert.NoError(t, a.Go(f.funcCtxC))

	aa := a.ForkChild()
	assert.NoError(t, aa.Go(f.funcCtxA))
	assert.NoError(t, aa.Go(f.funcCtxA))
	assert.NoError(t, aa.Go(f.funcCtxC, f.funcResetC))

	g.Wait()
	assert.EqualValues(t, 0, f.a)
	assert.EqualValues(t, 2, f.b)
	assert.EqualValues(t, 0, f.c)
}

func TestGroupRollback_3(t *testing.T) {
	f := class{}

	g := NewGroup()
	g.WithContext(context.TODO())
	assert.NoError(t, g.Go(f.funcCtxA))
	assert.NoError(t, g.Go(f.funcCtxA))
	assert.NoError(t, g.Go(f.funcCtxB, f.funcResetB))
	assert.NoError(t, g.Go(f.funcCtxC))
	assert.NoError(t, g.Go(func(ctx context.Context) error {
		return f.funcComplete(ctx, g.cancel)
	}))

	A := g.ForkChild()
	assert.NoError(t, A.Go(f.funcCtxA))
	assert.NoError(t, A.Go(f.funcCtxA))
	assert.NoError(t, A.Go(f.funcCtxB, f.funcResetB))
	assert.NoError(t, A.Go(f.funcCtxC))

	a := A.ForkChild()
	assert.NoError(t, a.Go(f.funcCtxA))
	assert.NoError(t, a.Go(f.funcCtxA))
	assert.NoError(t, a.Go(f.funcCtxC, f.funcResetC))
	assert.NoError(t, a.Go(f.funcTimeOut))

	aa := a.ForkChild()
	assert.NoError(t, aa.Go(f.funcCtxA))
	assert.NoError(t, aa.Go(f.funcCtxA, f.funcResetA))
	assert.NoError(t, aa.Go(f.funcCtxC))

	<-time.After(time.Millisecond * 300)
	g.Wait()

	assert.EqualValues(t, 0, f.a)
	assert.EqualValues(t, 2, f.b)
	assert.EqualValues(t, 0, f.c)

}

func TestGroupRollback_4(t *testing.T) {
	f := class{}

	g := NewGroup()
	g.WithContext(context.TODO())
	assert.NoError(t, g.Go(f.funcCtxA))
	assert.NoError(t, g.Go(f.funcCtxB))
	assert.NoError(t, g.Go(f.funcCtxC))
	assert.NoError(t, g.Go(f.funcTimeOut))

	A := g.ForkChild()
	A.DiscardedContext()
	assert.NoError(t, A.Go(f.funcA))
	assert.NoError(t, A.Go(f.funcB))
	assert.NoError(t, A.Go(f.funcC))

	a := A.ForkChild()
	a.WithContext(context.TODO())
	assert.NoError(t, a.Go(f.funcCtxA))
	assert.NoError(t, a.Go(f.funcCtxA, f.funcResetA))
	assert.NoError(t, a.Go(f.funcCtxC, f.funcResetC))

	time.AfterFunc(time.Millisecond*250, a.Close)
	g.Wait()
	assert.EqualValues(t, 0, f.a)
	assert.EqualValues(t, 2, f.b)
	assert.EqualValues(t, 0, f.c)
}

var count uint64
var m sync.Mutex

func timeAdd() error {
	<-time.After(time.Millisecond * 10)
	m.Lock()
	defer m.Unlock()
	count++
	return nil
}

func TestGroupMax(t *testing.T) {
	g := NewGroup()
	g.SetMaxGoroutine(50)
	_i := 0
	time.AfterFunc(time.Millisecond*5, func() {
		assert.EqualValues(t, uint64(50), g.GetGoroutineNum())
	})
	time.AfterFunc(time.Millisecond*15, func() {
		assert.EqualValues(t, uint64(100), g.GetGoroutineNum())
	})
	time.AfterFunc(time.Millisecond*25, func() {
		assert.EqualValues(t, uint64(150), g.GetGoroutineNum())
	})
	for _i < 1000 {
		g.Go(timeAdd)
		_i++
	}
	g.Wait()
	assert.EqualValues(t, uint64(1000), g.GetGoroutineNum())
	assert.EqualValues(t, uint64(1000), count)
}
