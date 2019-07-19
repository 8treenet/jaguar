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
    //session 自定义的连接插件
    session := plugins.NewSession()
    //连接加入插件， 后续可通过依赖注入获取
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
	Attach(plugin interface{})
	AttachImpl(impl string, plugin interface{})
	Close()
	RemoteAddr() net.Addr
	Push(pushHandle Encode) error
}

```