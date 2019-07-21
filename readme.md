# jaguar
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/8treenet/gotree/blob/master/LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/8treenet/tcp)](https://goreportcard.com/report/github.com/8treenet/tcp) [![Build Status](https://travis-ci.org/8treenet/gotree.svg?branch=master)](https://travis-ci.org/8treenet/gotree) [![GoDoc](https://godoc.org/github.com/8treenet/gotree?status.svg)](https://godoc.org/github.com/8treenet/gotree) [![QQ群](https://img.shields.io/:QQ%E7%BE%A4-602434016-blue.svg)](https://github.com/8treenet/jaguar) 

###### jaguar 是一个可扩展、高效的网络库。

## Overview
- Tcp Server
- Tcp Connect
- Request Handle
- Push Handle
- Middleware
- Inversion of Control
- Examples


## Tcp Server
```go

type TcpServer interface {
    Listen(*Opt)
    Accept(func(*TcpConn, *Middleware))
}

//创建 server
server := jaguar.NewServer()

//创建配置
opt := &jaguar.Opt{
    Addr: "0.0.0.0:9000",       //绑定地址和端口
    PacketMaximum: 6000,       //connect 可接收的最大包体字节，超过该字节主动断开连接。
    PacketHeadLen: 4,           //包头长度
    IdleCheckFrequency: time.Second * 120, //心跳检测
    ByteOrder: binary.BigEndian //网络字节序
}

// 新连接回调
// conn : 新连接
// middleware : 中间件
server.Accept(func(conn *jaguar.TcpConn, middleware *jaguar.Middleware) {
    //session 自定义的插件
    session := plugins.NewSession()
    //连接附加插件
    conn.Attach(session)
    //使用自定义的 session.CloseEvent 注册连接关闭事件
    middleware.Closed(session.CloseEvent)
})

//开启监听
server.Listen(opt)
```

## Tcp Connect
```go

type TcpConn interface {
    //附加插件，依赖注入获取 {p plugin `inject:""`}
    Attach(plugin interface{})
    //附加插件接口形式，依赖注入获取 {p inteface `inject:"impl"`}
    AttachImpl(impl string, plugin interface{})
    //关闭连接
    Close()
    //获取远程地址
    RemoteAddr() net.Addr
    //推送 推送处理器
    Push(pushHandle Encode) error
}
```


## Request Handle
```go
// 请求处理器，读写io数据和返回响应，继承该接口可使用。
type ReqHandle interface {
    //读取流数据
    ReadStream(...interface{}) error
    //读取流数据中的字符串
    ReadStreamByString(int, *string) error
    //写入流数据
    WriteStream(...interface{})
    //回执响应
    Respone() error
}

func init() {
    // 注册请求处理器 对应请求数据包的协议id 101
    // 请求处理器chat必须实现 Execute 方法
    jaguar.AddRequest(101, new(chat))
}

//定义一个聊天请求处理器
type chat struct {
    //继承 ReqHandle
    jaguar.ReqHandle `inject:"req_handle"`
    //通过依赖注入获取插件-实体方式
    Session *plugins.Session `inject:""`
    //使用 jaguar.TcpConn 插件, 该插件负责连接相关
    Conn jaguar.TcpConn `inject:"tcp_conn"`
}

// Execute - 必须实现
func (c *chat) Execute() {
    //c.ReadStream()
    //c.WriteStream()
    //c.Respone()
}
```

## Push Handle
```go
// 推送处理器，写io数据，继承该接口可使用。
type PushHandle interface {
    //Write byte stream data
    WriteStream(values ...interface{})
}

// 定义个推送处理器
type chat struct {
    jaguar.PushHandle `inject:"push_handle"`
    //要推送的数据列 三项
    sender            string
    content           string
    row               uint32
}

// ProtocolId - 必须实现
func (c *chat) ProtocolId() uint16 {
    //推送数据包的协议id
    //jaguar.TcpConn 插件会调用 chat.ProtocolId()
    return 300
}

// Encode() - 必须实现
func (c *chat) Encode() {
    //jaguar.TcpConn 插件会调用 chat.Encode()
    c.WriteStream(c.row)
    c.WriteStream(uint8(len(c.sender)), c.sender)
    c.WriteStream(uint16(len(c.content)), c.content)
}

TcpConn.Push(&chat{sender:"lucifer", content:"fuck"})
```


## Middleware
```go
// 拦截器
server.Accept(func(conn jaguar.TcpConn, middleware *jaguar.Middleware) {
    middleware.Closed(f func(){
        // Closed - Close registration for callbacks
    })

    middleware.Recover(f func(e error,s string){
        // Recover - Request a panic in the code
    })

    middleware.Reader(func(id uint16, b *bytes.Buffer) *bytes.Buffer {
        // Reader - Read data
        return b
    })

    middleware.Writer(func(id uint16, b *bytes.Buffer) *bytes.Buffer {
        // Writer - Write data
        return b
    })

    middleware.Request(func(id uint16, reqHandle interface{}) {
        //Request - Callback at request
    })

    middleware.Respone(func(id uint16, reqHandle interface{}) {
        //Respone - Callback when requesting a reply
    })

    middleware.Push(func(uint16, interface{}){
        //Push - Callbacks on push
    })
})
```

## Examples
##### 一个聊天室示例
```sh
# 启动示例程序
$ cd jaguar/chat_examples
# 启动server
$ go run server/main.go 

# 开启新窗口启动客户端 1
$ command + t
$ go run mock_client/main.go 

# 开启新窗口启动客户端 2
$ command + t
$ go run mock_client/main.go 
```

```
报文格式, 可变长数据需要指明 [长度, 数据]
+-------------------+--------------+---------------------------------------------------+
|               4 Bytes            |  2 Bytes  |       N Bytes                         |
+-------------------+--------------+---------------------------------------------------+
|<=         length of body       =>|     id    | <======= data =======================>|
|<============= header ===========>|<==================== body =======================>|

协议 ：请求登录
id : 100 (2 Bytes)
token :[] (1 Bytes, N Bytes)

协议 ：请求登录回执
id : 100 (2 Bytes)
ok : 1 (1 Bytes)
uid : [](4 Bytes)
uname : [](1 Bytes, N Bytes)


协议 ：请求聊天消息
id : 101 (2 Bytes)
row : [] (4 Bytes)
content : [](2 Bytes, N Bytes)

协议 ：请求聊天消息回执
id : 101 (2 Bytes)
ok : 1 (1 Bytes)

协议 ：聊天信息推送
id : 300 (2 Bytes)
sender : [] (1 Bytes, N Bytes)
content : [](2 Bytes, N Bytes)
```