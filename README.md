# PoolGroup
# 一个人性化的协程管理包，适用于高并发量，简单，复杂并发业务场景。

> 安装 go get github.com/XeiTongXueFlyMe/poolgroup

> 使用 import “github.com/XeiTongXueFlyMe/poolgroup”

## PoolGroup包，分为group and pool。

> group 解决复杂的并发逻辑

> pool 解决高并发量

> group + pool 能解决带有复杂逻辑的高并发 


### 优雅的使用并发
> 示例(独立组)


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
* 协程业务回滚
* 独立组   =>  func( ) error
* 上下文组   =>  func(ctx context.Context) error
* 派生树（父子关系，兄弟关系）
* 自由组合和派生。



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

### 协程业务回滚(上下文组)
> 1. 子组触发回滚，其父不回滚
> 2. 父组触发回滚,子组树全部产生回滚,其中不带上下文的独立组及其派生的子树不回滚
> 3. 在同一个group中，并发协程中某一个协程返回错误,或则panic时，所有协程执行业务回滚
> 4. 协程业务回滚 入口函数为 .(func() error)
```go
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
//out:
//成功写入golang实战.pdf
//PrintFileMeta:[{golang实战 .pdf 1566304752}]
//回滚被执行:[]
func main(){
    c := db{}
    g := group.NewGroup()
    g.WithContext(context.TODO())
    //g所有协程未返回err或者panic, c.DelAllFile()不会运行
    g.Go(c.AddFile, c.DelAllFile)
    g.Go(c.PrintFileMeta)
    
    g.Wait()
    
    return nil
}

```

### 关闭一个group
> 会触发协程业务回滚

```go
    g.Close()
```

### 读取派生树整个协程数量

```go
    g.GetGoroutineNum()
```

### 获取整个派生树的错误

> 可实时读取错误，并发安全

> g.wait()之后调用，获取本次执行整个派生树的错误
```go
    g.GetErrs()
```

### group 支持 .(func() error) 和.(func(ctx context.Context) error)协程运行入口,那么如何安全的向协程中带入参数呢？

*  不建议在ctx带入key&Value传参

> 下面实列将 a,b 参数带入协程
```go
func myPrintf(a, b string) error {
    fmt.Println(a, b)
    return nil
}
//out:
//m immm
//i immm
//h immm
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

```

### group上下文组支持Context，用于内部派生树，可用于SOA分布式架构，微服务架构等，,传递链路追踪消息，超时控制，特殊值传递等。

```go
func (t *calc) IncreasedCtx(ctx context.Context) error {
    for {
        time.Sleep(1 * time.Second)
        select {
        case <-ctx.Done():
        	return nil
        default:
        }
        t.m.Lock()
        t.value++
        t.m.Unlock()
    }
    return nil
}
func (t *calc) PrintValueCtx(ctx context.Context) error {
for {
    time.Sleep(1 * time.Second)
    select {
    case <-ctx.Done():
    	return nil
    default:
    }
    t.m.Lock()
    fmt.Println(t.value)
    t.m.Unlock()
    return nil
}
}
func main() {
    c := calc{value: 0}
    
    g := group.NewGroup()
    g.WithContext(context.TODO())
    ////10秒后g及其子组中协程全部退出。独立组节点及其子树除外
    //g.WithTimeout(context.TODO(), 10*time.Second)
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

### 自由组合和派生，需要注意什么
* group 分为两种： 独立组（.(func() error) ）上下文组（.(func(ctx context.Context) error)）
* 先调用g.WithContext()等组配置属性接口，在调用g.Go()。否则会panic
* 配置接口 g.WithContext()  g.WithTimeout()  g.DiscardedContext()
* g.Go() 正确调用方式， err : = g.Go(f) , 如果f入口函数格式错误，g.Go()会返回错误。如果你肯定f格式是正确的可以不用接收处理err
* 子组会继承父组的属性（独立组 or 上下文组）,配置接口可以改变这个属性
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
    
    aa := A.ForkChild()
    //...
    ab := A.ForkChild()
    ab.WithContext(context.TODO())
    //...
    bb : = B.ForkChild()
    bb.DiscardedContext()
    
    cc := bb.ForkChild()
    cc.WithTimeout(context.TODO(), 100*time.Millisecond)
    
    //直到所有的group退出，才退出
    g.Wait()
    return nil
}
```


### 回滚+自由组合和派生，让你复杂的业务变得简单
