package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/XeiTongXueFlyMe/poolgroup/group"
	"time"
)

//out: 123[err1 err2 err3]
func example_1() {

	g := group.NewGroup()
	g.Go(func() error {
		time.Sleep(100 * time.Millisecond)
		fmt.Print("1")
		return errors.New("err1")
	})
	g.Go(func() error {
		time.Sleep(200 * time.Millisecond)
		fmt.Print("2")
		return errors.New("err2")
	})
	g.Go(func() error {
		time.Sleep(300 * time.Millisecond)
		fmt.Print("3")
		return errors.New("err3")
	})

	g.Wait()
	fmt.Println(g.GetErrs())
}

func main() {
	ctx := context.TODO()
	cc, c := context.WithCancel(ctx)
	ccc := cc
	go func() {
		<-cc.Done()
		fmt.Print("1")
	}()
	go func() {
		<-ccc.Done()
		fmt.Print("2")
	}()

	c()
	time.Sleep(100 * time.Microsecond)
}
