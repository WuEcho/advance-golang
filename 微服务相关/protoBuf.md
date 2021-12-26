# protoBuf

## 1.简介
　　protoBuf(Google Protocol Buffer)是一种轻便高效的结构化数据存储格式，平台无关，语言无关，可拓展，可用于通信协议和数据存储等领域。

**数据交互格式比较**

- json: 一般的web项目中，最流行的主要还是json。
- xml: 在webservice中应用最为广泛，相对于json，它的数据更加冗余，需要成对的闭合标签。
- protobuf: 谷歌开源的一种数据格式，适合高性能，对响应速度有要求的数据传输场景。需要编码和解码。数据本身不可读。只能反序列之后读取。

**优缺点**

　　优点：体积小，简单。可自定义自己的数据结构，使用代码生成器生成的代码读写这个数据结构。可在无需重新部署程序的情况下更新数据结构。
　　缺点：数据本身不可读

## 2.安装
### 2.1 安装protoBuf

```
//下载 protoBuf
$ git clone https://github.com/protocolbuffers/protobuf.git

//安装依赖库
$ sudo apt-get install autoconf automake libtool curl make g++ unzip libffi- dev -y

//安装
$ cd protobuf/
$ ./autogen.sh 
$ ./configure 
$ make 
$ sudo make install 
$ sudo ldconfig # 刷新共享库 很重要的一步

#成功后需要使用命令测试 
$ protoc –h
```

### 2.2 获取proto包


```
#Go语言的proto API接口 
$ go get -v -u github.com/golang/protobuf/proto
```

### 2.3 安装protoc-gen-go插件

是一个go程序，编译后将可执行文件复制到/bin目录

```
#安装 
$ go get -v -u github.com/golang/protobuf/protoc-gen-go 
#编译 
$ cd $GOPATH/src/github.com/golang/protobuf/protoc-gen-go/ 
$ go build
#将生成的 protoc-gen-go可执行文件，放在/bin目录下 
$ sudo cp protoc-gen-go /bin/ 
```

## 3.protobuf 

### 3.1 语法
要想使用protobuf必须先定义proto文件。

#### 3.1.1 定义一个消息类型

```
syntax = "proto3";

message PandaRequest {
   string name = 1;
   int32 height = 2;
   repeated int32 weight = 3;
}
```
　　PandaRequest消息格式有三个字段，每个字段都有名字和类型。文件第一行指定了正在使用proto3语法，如果没指定，编译器会使用proto2。这个指定语法行必须是文件的非空非注释的第一行。
　　**Repeated**关键字表示重复的在go中用切片进行代表。

####3.1.2 添加更多消息类型

　　在一个.proto文件中可以定义多个消息类型。
　　

```
syntax = "proto3";

message PandaRequest {
   string name = 1;
   int32 height = 2;
   repeated int32 weight = 3;
}

message PandaResponse { 
   ... 
}
```

#### 3.1.3 添加注释
向.proto文件添加注释，可以使用双斜杠(//)

#### 3.1.4 使用其他消息类型
可以将其他消息类型用作字段类型。例如：

```
message PersonInfo {
  repeated Person info = 1;
}

message Person {
   string name = 1;
   int32 height = 2;
   repeated int32 weight = 3;
}

```

#### 3.1.5 使用proto2消息类型
在proto3消息中导入proto2的消息类型也可以，反之亦然。proto2枚举不可以直接在proto3的标识符中使用。

#### 3.1.6 嵌套类型
可以在其他消息类型定义中使用消息类型并可以嵌套任意多层，如：

```
message PersonInfo {
  message Person {
    message xxx {
      ...
    }
    string name = 1;
    int32 height = 2;
    repeated int32 weight = 3;
  }

  repeated Person info = 1;
}

```
如果想在其父消息类型外重用这个消息类型，可通过PersonInfo.Person的形式使用，如：


```
message PersonMessage {
  PersonInfo.Person info = 1;
}
```


### 3.2.标准数据类型


| .proto  Type | notes | C++  | Python | Go |
| --- | --- | --- | --- | --- |
| double |  | double | float | float64 |
| float |  | float | float | float32 |
| int32 |  使用变长编码，对于负值的效率很低，如果你的域有可能有负值，请使用sint64替代 | int32 | int | int32 |
| uint32 | 使用变长编码 | uint32 | int/long | uint32 |
| uint64 | 使用变长编码 | uint64 | int/long | uint64 |
| sint32 | 使用变长编码，这些编码在负值时比int32高效的多 | int32 | int | int32 |
| sint64 | 使用变长编码，有符号的整型值。编码时比通常的int64高效。| int64 | int/long | int64 |
| fixed32 | 总是4个字节，如果数值总是比总是比228大的话，这个类型会比uint32高效。| uint32 | int | uint32 |
| fixed64 | 总是8个字节，如果数值总是比总是比256大的话，这个类型会比uint64高效。| uint64 | int/long | uint64 |
| sfixed32 | 总是4个字节 | int32 | int | int32 |
| sfixed64 | 总是4个字节 | int64 | int/long | int64 |
| bool |  | bool | bool | bool |
| string | 一个字符串必须是UTF-8编码或者7-bit ASCII编码的文本。| string | str/unicode | string |
| bytes | 可能包含任意顺序的字节数据。 | string | str | []byte |


## 4.定义服务
　　若想将消息类型用在RPC系统中，可以在.proto文件中定义一个RPC服务接口，protocol buffer编译器将会根据选择不同的语言生成器生成服务接口代码及存根。如，想要定义一个RPC服务并具有一个方法，该方法能够接收SearchRequest并返回一个SearchResponse，此时可以在.proto文件中进行如下定义：
　　
```
service SearchService {
   //rpc 服务的函数名(传参) 返回 (返参)
   rpc Search (SearchRequest) returns (SearchResponse);
}
```



