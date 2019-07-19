package main

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"

	"github.com/8treenet/jaguar"
	"github.com/8treenet/jaguar/chat_examples/server/plugins"
	_ "github.com/8treenet/jaguar/chat_examples/server/request"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	server := jaguar.NewServer()
	opt := &jaguar.Opt{Addr: "0.0.0.0:9000", PacketMaximum: 6000, PacketHeadLen: 4, IdleCheckFrequency: time.Second * 120, ByteOrder: binary.BigEndian}

	//新连接处理
	server.Accept(func(conn jaguar.TcpConn, middleware *jaguar.Middleware) {
		fmt.Println("Access to a new connection :", conn.RemoteAddr().String())
		session := plugins.NewSession()
		//附加插件
		conn.Attach(session)
		//附加插件接口方式
		conn.AttachImpl("impl_example", session)
		middleware.Closed(session.CloseEvent)

		// middleware.Recover(session.Recover)
		// middleware.Reader(session.Reader)
		// middleware.Writer(session.Writer)
		// middleware.Request(session.Request)
		// middleware.Respone(session.Respone)
		// middleware.Push(session.Push)
	})

	fmt.Println("Listen :", *opt)
	//监听启动
	server.Listen(opt)
}
