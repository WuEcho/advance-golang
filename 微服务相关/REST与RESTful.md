# REST与RESTful

## 1.REST
REST -- Resource Representational State Transfer

- Resource:资源，即数据
- Representational:某种表达形式，比如JSON,XML
- State Transfer:状态变化。通过HTTP动词实现

   REST实际上是一种设计风格(对于api的设计),是面向资源的，资源是通过URL进行暴露的。所以REST就是选择通过http协议和url,利用client/server model对资源进行CRUD(Create/Read/Update/Delete)增删改查操作。换句话说，看url就知道要什么，看http method就知道干什么，看http status code就知道结果如何。
   
   作为一种架构，其提出了一系列架构级约束。这些约束有：
  
  - 1.使用客户/服务器模型。客户和服务器之间通过一个统一的接口来互相通讯。
  - 2.层次化的系统。在一个REST系统中，客户端并不会固定的与一个服务器打交道。
  - 3.无状态。在一个REST系统中，客户端并不会保存有关客户的任何状态。客户端自身负责用户状态的维持，并在每次发送请求时都需要提供足够的信息。
  - 4.可缓存。REST系统需要能够恰当地缓存请求，以尽量减少服务端和客户端之间的信息传输，提高性能。
  - 5.统一的接口。一个REST系统需要使用一个统一的接口来完成子系统之间以及服务与用户之间的交互。 

  如果一个系统满足上面所有的约束，那么该系统被称为RESTful的。

更多详细可参考阅读：[REST简介](https://www.cnblogs.com/loveis715/p/4669091.html)


