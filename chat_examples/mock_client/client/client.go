package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"
)

func NewMockConn(addr string, readcall func(uint16, *bytes.Buffer)) *TcpConn {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	ts := new(TcpConn)
	ts.Conn = conn
	ts.writeChannel = make(chan []byte, 4096)
	ts.readcall = readcall
	return ts
}

type TcpConn struct {
	net.Conn
	writeChannel chan []byte
	readcall     func(uint16, *bytes.Buffer)
}

func (self *TcpConn) Start() {
	go self.write()
	self.read()
}

func (self *TcpConn) read() {
	readStatus := 1
	bodyLen := uint32(0)
	packet := new(bytes.Buffer)
	for {
		self.SetReadDeadline(time.Now().Add(time.Second * 60))
		packet.Reset()
		var data []byte
		var trialLen int
		if readStatus == 1 {
			trialLen = 4
		} else {
			trialLen = int(bodyLen)
		}

		for trialLen > 0 {
			data = make([]byte, trialLen)
			len, derr := self.Read(data)
			if len == 0 || derr != nil {
				self.Conn.Close()
				self.readBreak()
				return
			}

			packet.Write(data[:len])
			trialLen = trialLen - len
		}

		if readStatus == 1 {
			readStatus = 2
			bodyLen = self.readBodyLen(packet.Bytes())
		} else {
			readStatus = 1
			packetHandle := new(bytes.Buffer)
			packetHandle.Write(packet.Bytes())
			go self.routeHandle(packetHandle)
		}
	}
}

// write
func (self *TcpConn) write() {
	for {
		data := <-self.getWriteChannel()
		head := len(data)
		if head == 0 {
			return
		}

		headData := self.uintToBytes(uint32(head))
		self.SetWriteDeadline(time.Now().Add(time.Second * 15))
		if _, err := self.Write(headData); err != nil {
			return
		}
		if _, err := self.Write(data); err != nil {
			return
		}
	}
}

func (self *TcpConn) readBodyLen(head []byte) uint32 {
	buffer := bytes.NewBuffer(head)
	var number uint32
	binary.Read(buffer, binary.BigEndian, &number)
	return number
}

// break
func (self *TcpConn) readBreak() {
	fmt.Println("Close")
	os.Exit(1)
}

//Close 关闭
func (self *TcpConn) Close() {
	self.getWriteChannel() <- make([]byte, 0)
	self.Conn.Close()
}

//GetWriteChannel 获取写入管道
func (self *TcpConn) getWriteChannel() chan []byte {
	return self.writeChannel
}

func (self *TcpConn) Send(data []byte) {
	self.getWriteChannel() <- data
}

func (self *TcpConn) routeHandle(buffer *bytes.Buffer) {
	var id uint16 = 0
	binary.Read(buffer, binary.BigEndian, &id)
	self.readcall(id, buffer)
}

func (self *TcpConn) uintToBytes(number uint32) []byte {
	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, number)
	return buffer
}
