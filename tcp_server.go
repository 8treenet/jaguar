package jaguar

import (
	"encoding/binary"
	"net"
	"time"
)

var (
	_opt *Opt
)

type Opt struct {
	Addr               string           // The listen network address
	PacketMaximum      uint32           // The maximum byte of a network packet, default 6000
	PacketHeadLen      int8             // Header length of network packet, In 1, 2, 4, 8 byte, default 4
	IdleCheckFrequency time.Duration    // Check for idle connection times, during which no data access will be closed, defailt 120 sec.
	ByteOrder          binary.ByteOrder // The default is binary.BigEndian
}

type TcpServer interface {
	Listen(*Opt)
	Accept(func(TcpConn, *Middleware))
}

func NewServer() TcpServer {
	ts := new(tcpServer)
	return ts
}

type tcpServer struct {
	socket       net.Listener
	beforeAccept func(tc TcpConn, hook *Middleware)
}

// accept
func (ts *tcpServer) accept() {
	for {
		connect, err := ts.socket.Accept()
		if err != nil {
			return
		}
		hook := new(Middleware)
		client := newConn(connect, hook)
		ts.beforeAccept(client, hook)
		client.attachDi()
		client.start()
	}
}

func (ts *tcpServer) Accept(call func(tc TcpConn, hook *Middleware)) {
	ts.beforeAccept = call
	return
}

// Listen
func (ts *tcpServer) Listen(opt *Opt) {
	if !InSlice([]int8{1, 2, 4, 8}, opt.PacketHeadLen) {
		panic("HeadLen error")
	}
	_opt = opt
	var err error
	ts.socket, err = net.Listen("tcp", _opt.Addr)
	if err != nil {
		panic(err.Error())
	}
	ts.accept()
	return
}

// Close
func (ts *tcpServer) Close(args ...interface{}) {
	ts.socket.Close()
}

func (ts *Opt) init() {
	if ts.PacketMaximum == 0 {
		ts.PacketMaximum = 6000
	}
	if ts.PacketHeadLen == 0 {
		ts.PacketHeadLen = 4
	}
	if ts.IdleCheckFrequency == 0 {
		ts.IdleCheckFrequency = time.Second * 120
	}
	if ts.ByteOrder == nil {
		ts.ByteOrder = binary.BigEndian
	}
}
