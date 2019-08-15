# PoolGroup
# 一个人性化的协程管理pkg

> 安装 go get github.com/XeiTongXueFlyMe/poolgroup

>使用 import “github.com/XeiTongXueFlyMe/poolgroup”

### 并发一组协程,直到所有协程出错或者退出,才退出
> 示例

```go
import “github.com/XeiTongXueFlyMe/poolgroup”

func main(){
    g := poolgroup.NewErrGroup()
    
    g.Go(func() error {
      time.Sleep(200 * time.Millisecond)
      return errors.New("err")
    })
    g.Go(func() error {
      time.Sleep(300 * time.Millisecond)
      return nil
    })
    
    g.Wait()
    
    //读取错误
    for _, err := range g.errs {
      fmt.Println(err)
    }
    //读取协程数
    g.GetGoroutineNum()
}
```

### 并发一组协程,有一个协程退出,立即退出所有协程
> 示例

```go
import “github.com/XeiTongXueFlyMe/poolgroup”

func main(){
    g := poolgroup.NewErrGroup()
    g.WithContext(context.Background())
    
    g.Go(func() error {
      time.Sleep(200 * time.Millisecond)
      return errors.New("err")
    })
    g.Go(func() error {
      time.Sleep(300 * time.Millisecond)
      return nil
    })
    
    g.Wait()
}
```

### 并发一组协程,有一个协程退出或则超时,立即退出所有协程
> 示例

```go
import “github.com/XeiTongXueFlyMe/poolgroup”

func main(){
    g := poolgroup.NewErrGroup()
    g.WithTimeout(context.Background(), 1000*150) //150ms
    
    g.Go(func() error {
      time.Sleep(200 * time.Millisecond)
      return errors.New("err")
    })
    g.Go(func() error {
      time.Sleep(300 * time.Millisecond)
      return nil
    })
    
    g.Wait()
}
```

> 测试用例 并发一组协程,直到所有协程出错或者退出,才退出
```go
func TestGroup(t *testing.T) {
	counts, m := 0, sync.Mutex{}

	g := NewErrGroup()
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
		return errors.New("err")
	})

	g.Wait()
	for _, err := range g.errs {
		assert.Error(t, err)
	}
	assert.EqualValues(t, 2, g.GetGoroutineNum())
	assert.EqualValues(t, 2, counts)
}
```



> 测试用例 并发一组协程,有一个协程退出,立即退出所有协程
```go
func TestGroupWithContext(t *testing.T) {
	counts, m := 0, sync.Mutex{}

	g := NewErrGroup()
	g.WithContext(context.Background())
	g.Go(func() error {
		time.Sleep(10 * time.Millisecond)
		m.Lock()
		defer m.Unlock()
		counts++
		return nil
	})
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
		return errors.New("err")
	})

	g.Wait()
	for _, err := range g.errs {
		assert.Error(t, err)
	}
	assert.EqualValues(t, 3, g.GetGoroutineNum())
	assert.EqualValues(t, 2, counts)
}
```

> 测试用例 并发一组协程,有一个协程退出或则超时,立即退出所有协程
```go
func TestGroupWithTimeout(t *testing.T) {
	counts, m := 0, sync.Mutex{}

	g := NewErrGroup()
	g.WithTimeout(context.Background(), 1000*150) //150ms
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
		return errors.New("err")
	})

	g.Wait()
	for _, err := range g.errs {
		assert.Error(t, err)
	}
	assert.EqualValues(t, 2, g.GetGoroutineNum())
	assert.EqualValues(t, 1, counts)
}
```