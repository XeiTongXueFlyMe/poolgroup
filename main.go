package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/XeiTongXueFlyMe/poolgroup/group"
	"sync"
	"time"
)

func example_1() {
	g := group.NewGroup()

	g.Go(func() error { return errors.New("hi, i am Task_1") })
	g.Go(func() error { return errors.New("hi, i am Task_2") })
	g.Go(func() error { return errors.New("hi, i am Task_3") })

	//阻塞，直到本组中所有的协程都安全的退出
	g.Wait()
}

func fPanic() error {
	panic("The err is unknown")
}

func example_2() {
	g := group.NewGroup()

	g.Go(fPanic)
	g.Go(func() error { return nil })
	g.Go(func() error { return errors.New("runtime err") })

	//阻塞，直到本组中所有的协程都安全的退出
	g.Wait()
	fmt.Println(g.GetErrs())
}

type calc struct {
	value int
	m     sync.Mutex
}

func (t *calc) Increased() error {
	t.m.Lock()
	defer t.m.Unlock()
	t.value++
	return nil
}
func (t *calc) PrintValue() error {
	t.m.Lock()
	defer t.m.Unlock()
	fmt.Println(t.value)
	return nil
}

func (t *calc) IncreasedCtx(ctx context.Context) error {

	for {
		time.Sleep(100 * time.Millisecond)
		select {
		case <-ctx.Done():
		default:
		}
		t.m.Lock()
		t.value++
		t.m.Unlock()
		if t.value > 2 {
			return nil
		}
	}
	return nil
}
func (t *calc) PrintValueCtx(ctx context.Context) error {
	for {
		time.Sleep(200 * time.Millisecond)
		select {
		case <-ctx.Done():
		default:
		}
		fmt.Println(t.value)
		return nil
	}
}

func example_3() {
	c := calc{value: 0}

	g := group.NewGroup()
	g.Go(c.Increased)
	g.Go(c.PrintValue)
	g.Go(func() error { return nil })

	g.Wait()
}
func example_4() {
	c := calc{value: 0}

	g := group.NewGroup()
	g.WithContext(context.TODO())
	//g.WithTimeout(context.TODO(), 200*time.Millisecond)
	g.Go(c.IncreasedCtx)
	g.Go(c.PrintValueCtx)

	g.Wait()
}

//func main() {
//	ctx := context.TODO()
//	cc, c := context.WithCancel(ctx)
//	ccc := cc
//	go func() {
//		<-cc.Done()
//		fmt.Print("1")
//	}()
//	go func() {
//		<-ccc.Done()
//		fmt.Print("2")
//	}()
//
//	c()
//	time.Sleep(100 * time.Microsecond)
//}

//var rollback chan int
//rollback = make(chan int, 3)
//rollback <- 2
//rollback <- 2
//rollback <- 2
//go func() {
//	for {
//		select {
//		case <-rollback:
//			fmt.Println("rollback")
//			continue
//		default:
//			fmt.Println("default")
//		}
//		break
//	}
//}()

func main() {
	example_1()
	example_2()
	example_3()
	example_4()
}
