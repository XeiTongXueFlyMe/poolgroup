package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/XeiTongXueFlyMe/poolgroup/group"
	"sync"
	"time"
)

func example_1() error {
	g := group.NewGroup()

	g.Go(func() error { return errors.New("hi, i am Task_1") })
	g.Go(func() error { return errors.New("hi, i am Task_2") })
	e := g.Go(func() error { return errors.New("hi, i am Task_3") })
	if e != nil {
		panic(e)
	}

	//阻塞，直到本组中所有的协程都安全的退出
	g.Wait()

	return nil
}

func fPanic() error {
	panic("The err is unknown")
}

func example_2() error {
	g := group.NewGroup()

	g.Go(fPanic)
	g.Go(func() error { return nil })
	g.Go(func() error { return errors.New("runtime err") })

	//阻塞，直到本组中所有的协程都安全的退出
	g.Wait()
	fmt.Println(g.GetErrs())

	return nil
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
}

//模拟一个函数运行时发生错误
func (t *calc) TimeOutErr(ctx context.Context) error {
	time.Sleep(100 * time.Millisecond)
	return errors.New("TimeOut")
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

func example_3() error {
	c := calc{value: 0}

	g := group.NewGroup()
	g.Go(c.Increased)
	g.Go(c.PrintValue)
	g.Go(func() error { return nil })

	g.Wait()

	return nil
}
func example_4() error {
	c := calc{value: 0}

	g := group.NewGroup()
	g.WithContext(context.TODO())
	//g.WithTimeout(context.TODO(), 200*time.Millisecond)
	g.Go(c.IncreasedCtx)
	g.Go(c.PrintValueCtx)

	g.Wait()

	return nil
}

func example_5() error {
	g := group.NewGroup()
	g.Go(func() error { return nil })

	A := g.ForkChild()
	A.Go(func() error { return nil })
	B := g.ForkChild()
	B.Go(func() error { return nil })
	C := g.ForkChild()
	C.Go(func() error { return nil })

	a := A.ForkChild()
	a.Go(func() error { return nil })
	b := A.ForkChild()
	b.Go(func() error { return nil })
	c := A.ForkChild()
	c.Go(func() error { return nil })

	g.Wait()
	return nil
}

func example_6() error {
	c := calc{value: 0}

	g := group.NewGroup()
	g.WithContext(context.TODO())
	g.Go(c.IncreasedCtx)
	g.Go(c.TimeOutErr)

	A := g.ForkChild()
	A.Go(c.IncreasedCtx)
	B := g.ForkChild()
	B.Go(c.IncreasedCtx)

	a := A.ForkChild()
	a.Go(c.IncreasedCtx)
	b := A.ForkChild()
	b.Go(c.IncreasedCtx)
	b.Go(c.IncreasedCtx)

	g.Wait()
	fmt.Println("所有协程全部退出")
	return nil
}

type metaData struct {
	Name       string
	Ext        string
	createTime int64
}
type db struct {
	file []metaData
}

func (db *db) AddFile(ctx context.Context) error {
	db.file = append(db.file, metaData{
		Name:       "golang实战",
		Ext:        ".pdf",
		createTime: time.Now().Unix(),
	})
	fmt.Println("成功写入golang实战.pdf")
	//模拟一个错误,由于不可抗拒原因本协程出现错误
	//panic("直接panic，也是可以的")
	return errors.New("某个步骤返回错误")
}

//DelAllFile is  rollback func
func (db *db) DelAllFile() error {
	db.file = []metaData{}
	fmt.Printf("回滚被执行:")
	fmt.Println(db.file)
	return nil
}
func (db *db) PrintFileMeta(ctx context.Context) error {
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("PrintFileMeta:")
	fmt.Println(db.file)
	return nil
}

//out : []
func example_7() error {
	c := db{}
	g := group.NewGroup()
	g.WithContext(context.TODO())
	g.Go(c.AddFile, c.DelAllFile)
	g.Go(c.PrintFileMeta)

	g.Wait()

	return nil
}

func myPrintf(a, b string) error {
	fmt.Println(a, b)
	return nil
}
func example_8() error {
	a := []string{"h", "i", "m"}
	b := "immm"

	g := group.NewGroup()
	for _, v := range a {
		value := v
		g.Go(func() error {
			return myPrintf(value, b)
		})
	}
	g.Wait()

	return nil
}

func main() {
	g := group.NewGroup()
	//g.Go(example_1)
	//g.Go(example_2)
	//g.Go(example_3)
	//g.Go(example_4)
	//g.Go(example_5)
	//g.Go(example_6)
	//g.Go(example_7)
	g.Go(example_8)
	g.GetErrs()

	g.Wait()
}
