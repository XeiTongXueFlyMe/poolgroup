package group

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestGroup(t *testing.T) {
	counts, m := 0, sync.Mutex{}

	g := NewGroup()
	g.Go(func() error {
		time.Sleep(100 * time.Millisecond)
		m.Lock()
		defer m.Unlock()
		counts++
		return errors.New("err")
	})
	g.Go(func() error {
		time.Sleep(200 * time.Millisecond)
		m.Lock()
		defer m.Unlock()
		counts++
		return nil
	})
	g.Go(func() error {
		time.Sleep(300 * time.Millisecond)
		m.Lock()
		defer m.Unlock()
		counts++
		return errors.New("err")
	})
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
	g.Go(func() error {
		time.Sleep(100 * time.Millisecond)
		m.Lock()
		defer m.Unlock()
		counts++
		return errors.New("err")
	})
	g.Go(func() error {
		time.Sleep(200 * time.Millisecond)
		m.Lock()
		defer m.Unlock()
		counts++
		panic("panic")
		return nil
	})
	g.Go(func() error {
		time.Sleep(300 * time.Millisecond)
		m.Lock()
		defer m.Unlock()
		counts++
		return errors.New("err")
	})
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
func (this *class) funcB() error {
	time.Sleep(100 * time.Millisecond)
	this.m.Lock()
	defer this.m.Unlock()
	this.b++

	return nil
}
func (this *class) funcC() error {
	time.Sleep(100 * time.Millisecond)
	this.m.Lock()
	defer this.m.Unlock()
	this.c++

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

func TestGroupWait(t *testing.T) {
	f := class{}

	g := NewGroup()
	g.Go(f.funcA)
	g.Go(f.funcB)
	g.Go(f.funcC)

	A := g.ForkChild()
	A.Go(f.funcB)
	A.Go(f.funcC)

	B := g.ForkChild()
	B.WithContext(context.TODO())
	B.Go(f.funcCtxA)
	B.Go(f.funcCtxC)
	B.Go(f.funcCtxC)
	B.Go(f.funcTimeOut)

	a := B.ForkChild()
	a.Go(f.funcCtxA)
	a.Go(f.funcCtxC)
	a.Go(f.funcCtxC)
	a.Go(f.funcTimeOut)

	b := A.ForkChild()
	b.Go(f.funcB)
	b.Go(f.funcC)

	g.Wait()
	assert.EqualValues(t, 3, f.a)
	assert.EqualValues(t, 3, f.b)
	assert.EqualValues(t, 7, f.c)
}

func TestGroupDiscardedContext(t *testing.T) {
	f := class{}

	g := NewGroup()
	g.WithContext(context.TODO())
	g.Go(f.funcCtxA)
	g.Go(f.funcTimeOut)

	A := g.ForkChild()
	A.DiscardedContext()
	A.Go(f.funcA)

	g.Wait()
	fmt.Println(g.GetErrs())

	assert.EqualValues(t, 3, g.GetGoroutineNum())
	assert.EqualValues(t, 2, f.a)
}

//func TestGroupWithTimeout(t *testing.T) {
//	counts, m := 0, sync.Mutex{}
//
//	g := NewErrGroup()
//	g.WithTimeout(context.Background(), 1000*150) //150ms
//	g.Go(func() error {
//		time.Sleep(100 * time.Millisecond)
//		m.Lock()
//		defer m.Unlock()
//		counts++
//		return errors.New("err")
//	})
//	g.Go(func() error {
//		time.Sleep(100 * time.Millisecond)
//		m.Lock()
//		defer m.Unlock()
//		counts++
//		return errors.New("err")
//	})
//	g.Go(func() error {
//		time.Sleep(200 * time.Millisecond)
//		m.Lock()
//		defer m.Unlock()
//		counts++
//		return errors.New("err")
//	})
//
//	g.Wait()
//	for _, err := range g.errs {
//		assert.Error(t, err)
//	}
//	assert.EqualValues(t, 3, g.GetGoroutineNum())
//	assert.EqualValues(t, 2, counts)
//}
//
//func TestGroupWithContext(t *testing.T) {
//	counts, m := 0, sync.Mutex{}
//
//	g := NewErrGroup()
//	g.WithContext(context.Background())
//	g.Go(func() error {
//		time.Sleep(10 * time.Millisecond)
//		m.Lock()
//		defer m.Unlock()
//		counts++
//		return nil
//	})
//	g.Go(func() error {
//		time.Sleep(100 * time.Millisecond)
//		m.Lock()
//		defer m.Unlock()
//		counts++
//		return errors.New("err")
//	})
//	g.Go(func() error {
//		time.Sleep(200 * time.Millisecond)
//		m.Lock()
//		defer m.Unlock()
//		counts++
//		return errors.New("err")
//	})
//
//	g.Wait()
//	for _, err := range g.errs {
//		assert.Error(t, err)
//	}
//	assert.EqualValues(t, 3, g.GetGoroutineNum())
//	assert.EqualValues(t, 2, counts)
//}
