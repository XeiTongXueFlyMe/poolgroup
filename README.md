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
