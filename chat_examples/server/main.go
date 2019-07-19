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
	opt := &jaguar.Opt{Addr: "0.0.0.0:9000", PacketMaximum: 60000, PacketHeadLen: 4, IdleCheckFrequency: time.Second * 120, ByteOrder: binary.BigEndian}
	server.Accept(func(tcp *jaguar.TcpConn, middleware *jaguar.Middleware) {
		fmt.Println("Access to a new connection :", tcp.RemoteAddr().String())
		session := plugins.NewSession()
		tcp.Attach(session)
		middleware.Closed(session.CloseEvent)
	})

	fmt.Println("Listen :", *opt)
	server.Listen(opt)
}
