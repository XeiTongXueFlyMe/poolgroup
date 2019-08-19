# PoolGroup
# 一个人性化的协程管理包，适用于高并发量，简单，复杂并发逻辑场景。

> 安装 go get github.com/XeiTongXueFlyMe/poolgroup

> 使用 import “github.com/XeiTongXueFlyMe/poolgroup”

## PoolGroup包，分为group and pool。

> group 解决复杂的并发逻辑

> pool 解决高并发量

> group + pool 能解决带有复杂逻辑的高并发 


### 优雅的使用并发
> 示例

```go
import "github.com/XeiTongXueFlyMe/poolgroup/group"

func main(){
	g := group.NewGroup()

	g.Go(func() error { return errors.New("hi, i am Task_1") })
	g.Go(func() error { return errors.New("hi, i am Task_2") })
	g.Go(func() error { return errors.New("hi, i am Task_3") })

	//阻塞，直到本组中所有的协程都安全的退出
	g.Wait()
}

```

### group的特性

* 简单
* 轻量级
* panic安全
* 独立组 ( func( ) error )
* 上下文组( func(ctx context.Context) error )
* 自由组合和派生。
* 派生树（父子关系，兄弟关系）
* 协程业务回滚
    > 1. 子组触发回滚，其父不回滚
    > 2. 父组触发回滚,子组树全部产生回滚,其中不带上下文的独立组及其派生的子树不回滚
    > 3. 在同一个group中，并发协程中某一个协程返回错误,或则panic时，所有协程执行业务回滚

### pool的特性

## PoolGroup概念图

## group功能探索

### panic安全

> 协程抛出panic,整个组安全运行，group会将panic写入其 errs

```go
func fPanic() error {
	panic("The err is unknown")
}
//out: [runtime err The err is unknown]
func main(){
	g := group.NewGroup()

	g.Go(fPanic)
	g.Go(func() error { return nil })
	g.Go(func() error { return errors.New("runtime err") })

	g.Wait()
	fmt.Println(g.GetErrs())
}

```

### group 支持 .(func() error) 和.(func(ctx context.Context) error)协程运行入口,那么如何安全的向协程中带入参数呢？

*  不建议在ctx带入key&Value传参

```go
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

func main() {
	c := calc{value: 0}

	g := group.NewGroup()
	g.Go(c.Increased)
	g.Go(c.PrintValue)
	g.Go(func() error { return nil })

	g.Wait()
}

```

### group上下文组支持Context，由于内部派生树，可用于SOA分布式架构，微服务架构等，,传递链路追踪消息，超时控制，特殊值传递等。

```go
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
func main() {
	c := calc{value: 0}

	g := group.NewGroup()
	g.WithContext(context.TODO())
	//g.WithTimeout(context.TODO(), 200*time.Millisecond)
	g.Go(c.IncreasedCtx)
	g.Go(c.PrintValueCtx)

	g.Wait()
}

```

### 如何创建派生树，子group全部退出，父group才退出。group中任何一个协程返回错误，或则panic，其他协程，其他group照样运行
> 一个简单的派生树

> g 
>> A
>>> a

>>> b

>>> c

>> B 

>> C


```go
func main() {
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
    
    //直到所有的group退出，才退出
    g.Wait()
}
```

### 如何在派生树中，创建父子关系，像线程一样，父亲down机，其子线程停止运行。
> 下面代码：
> group中任何一个协程返回错误，或则panic，本group所有协程退出，其子树全部退出
```go
//模拟一个协程运行时发生错误
func (t *calc) TimeOutErr(ctx context.Context) error {
    time.Sleep(100 * time.Millisecond)
    return errors.New("TimeOut")
}

//out : 所有协程全部退出
func main() {
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
    
    //直到所有的group退出，才退出
    g.Wait()
    fmt.Println("所有协程全部退出")
    return nil
}
```

### 如何在派生树中，创建父子关系后，希望父down机，某个子及其子树不受影响。
> 如下示例，g组中某个协程返回错误，或则panic， B组及其子树全部退出，但是A组及其子树（a,b）不退出（除非自己安全退出）
```go
func main() {
    c := calc{value: 0}
    
    g := group.NewGroup()
    g.WithContext(context.TODO())
    //...
    
    A := g.ForkChild()
    A.DiscardedContext()
    //...
    B := g.ForkChild()
    //...
    
    a := A.ForkChild()
    //...
    b := A.ForkChild()
    //...
    
    //直到所有的group退出，才退出
    g.Wait()
    return nil
}
```
