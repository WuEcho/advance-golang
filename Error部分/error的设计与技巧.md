# Error的设计与技巧

总结与反思： 1.在日常开发过程中应对频繁的打日志没有寻找更改好的办法

## 1.Error 
go error就是一个普通的接口，普通的值

```
http://golang.org/pkg/builtin/#error

type error interface {
    Error()  string
}

```
经常使用**errors.New()**来返回一个error对象

```
http://golang.org/src/pkg/errors/errors.go

type errorString struct {
     s string
}

func (e *errorString) Error() string {
     return e.s
}
```

## 2. Error Type
### 2.1 Sentinel Error

　　预定义的特定错误，例如**io.EOF**等一般这种错误需要特定的值进行判等。`if err == ErrSomething(...)`。相对而言，Sentinel error是最不灵活的错误处理策略，因此调用方必须使用==将结果与预先声明的值进行比较。当想携带更多错误信息的上下文，将是个问题。

**不依赖检查error.Error的输出。**：Error方法存在于error接口主要用于调试，而不是编程。

Sentinel Error的一些弊端：1.成为API公共部分 2.在两个互不关联的包之间创建了依赖

因此尽可能避免使用Sentinel Error

### 2.2 Error types

Error type是实现了error接口的自定义类型。

```
type MyError struct {
  Msg      string
  File     string
  Line     int
}

func (e *MyError) Error() string {
  return fmt.Sprintf("%s:%d:%s",e.File,e.line,e.Msg)
}
```
因为MyError是一个type,因此可以使用类型断言进行类型判断

```
func main() {
   err := &MyError{"something happend","xxx.go",123}
   switch err := err.(type){
   case nil:
   
   case *MyError:
        fmt.Println("err occurred on line:",err.Line)
   default:
   }
}

```
　　这类错误可以携带更多的报错信息，**os.PathError**是此类型的示例。此种类型需要调用者使用类型断言和类型switch,还要让自定义的error变为public。

### 2.3 Opaque errors

　　不透明错误是一种非常灵活的错误处理策略，它要求代码和调用者之间的耦合最少。虽然外接调用者知道发生了错误，但是没有能力看到错误的内部。作为调用者，关于操作的结果，知道的只有其起作用了没有。


```
package net 

type Error interface {
  error
  Timeout() bool
  Temporary() bool
}

if nerr,ok := err.(net.Error);ok&&nerr.Temporary(){
  // do something
}
```
 
## 3.wrap error

　　日志记录与错误无关，且对调试没有帮助的信息应视为噪声。记录的原因是因为某些东西失败了，而日志包含了答案。在程序中，对于错误需要第一时间进行处理，因此可能产生大量的卫戍语言进行错误的判断操作。也可能大量底层代码在错误发生后将错误输出出日志。因此，随着项目的更加庞大，日志信息会越来会多，对排查错误并不友好。
　　为了既尽可能减少卫戍语言的出现，而且提供更多的错误上下文信息，可以借助`github.com/pkg/errors`这个包提供的一些方法可以即保存原始信息，也可以携带一些必要的上下文信息。
　　可以使用errors.Wrap或者errors.Wrapf保存堆栈信息。使用errors.Cause获取root error,再进行和sentinel error判定。

## 4.Go 1.13 error

　　有时需要自定义错误类型包裹底层错误，以**QueryError**为例：
　　
```
type QueryError struct {
    Query  string
    Err    error
}
```

　　可以通过查看QueryError值以底层错误作出决策：　
　　
```
if e,ok := err.(*QueryError);ok&& e.Err == ErrPermission {
  // do something
} 
```

　　在Go1.13中为errors和fmt标准库提供了新特性，以简化错误处理。最重要的是：包含另一个错误的error可以实现返回底层错误的**Unwrap**方法。
　　
```
func (q *QueryError) Unwrap() error {
   return q.Err
}
```

　　Go1.13提供两个新函数用于检测错误：**Is**和**As**。

```
//if err == ErrNotFound {...}
if errors.Is(err,ErrNotFound) {
   //do something
}

//if e,ok := err.(*QueryError);ok {...}
var e *QueryError
if errors.As(err,&QueryError){
  // do something
} 
```

　　在Go1.13中fmt.Errorf支持新的谓词**%w**,用**%w**包装的错误可用于errors.Is以及errors.As。

```
err := fmt.Errorf("xxx %w",ErrPermission)

if errors.Is(err,ErrPermission)
```

