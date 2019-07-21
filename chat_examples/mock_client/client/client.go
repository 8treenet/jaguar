package client

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"reflect"
	"time"
)

var (
	PacketHeaderLength = 2
	ByteOrder     = binary.BigEndian
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
	status := 1
	bodyLen := 0
	packet := new(bytes.Buffer)
	for {
		self.SetReadDeadline(time.Now().Add(time.Second * 60))
		packet.Reset()
		var data []byte
		var trialLen int
		if status == 1 {
			trialLen = PacketHeaderLength
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

		if status == 1 {
			status = 2
			bodyLen = self.packetLenToInt(packet.Bytes())
		} else {
			status = 1
			packetHandle := new(bytes.Buffer)
			packetHandle.Write(packet.Bytes())
			go self.routeHandle(packetHandle)
		}
	}
}

// write
func (self *TcpConn) write() {
	for {
		data := <-self.writeChannel
		head := len(data)
		if head == 0 {
			return
		}

		headData := self.packetLenToByte(head)
		self.SetWriteDeadline(time.Now().Add(time.Second * 15))
		if _, err := self.Write(headData); err != nil {
			return
		}
		if _, err := self.Write(data); err != nil {
			return
		}
	}
}

// packetLenToInt
func (tc *TcpConn) packetLenToInt(head []byte) int {
	buffer := bytes.NewBuffer(head)
	switch len(head) {
	case 1:
		var x uint8
		binary.Read(buffer, binary.BigEndian, &x)
		return int(x)
	case 2:
		var x uint16
		binary.Read(buffer, binary.BigEndian, &x)
		return int(x)
	case 4:
		var x uint32
		binary.Read(buffer, binary.BigEndian, &x)
		return int(x)
	case 8:
		var x uint64
		binary.Read(buffer, binary.BigEndian, &x)
		return int(x)
	}

	tc.Conn.Close()
	return 0
}

// packetSize
func (tc *TcpConn) packetLenToByte(plen int) []byte {
	switch PacketHeaderLength {
	case 1:
		data, err := toBytes(uint8(plen))
		if err == nil {
			return data
		}
	case 2:
		data, err := toBytes(uint16(plen))
		if err == nil {
			return data
		}
	case 4:
		data, err := toBytes(uint32(plen))
		if err == nil {
			return data
		}
	case 8:
		data, err := toBytes(uint64(plen))
		if err == nil {
			return data
		}
	}

	return []byte{}
}

// break
func (self *TcpConn) readBreak() {
	fmt.Println("Close")
	os.Exit(1)
}

//Close 关闭
func (self *TcpConn) Close() {
	self.writeChannel <- make([]byte, 0)
	self.Conn.Close()
}

func (self *TcpConn) Send(data []byte) {
	self.writeChannel <- data
}

func (self *TcpConn) routeHandle(buffer *bytes.Buffer) {
	var id uint16 = 0
	binary.Read(buffer, binary.BigEndian, &id)
	self.readcall(id, buffer)
}

func toBytes(dest interface{}) (result []byte, e error) {
	switch data := dest.(type) {
	case string:
		result = []byte(data)
	case []byte:
		result = []byte(data)
	case int8:
		result = append(result, byte(data))
	case uint8:
		result = append(result, byte(data))
	case uint16:
		result = make([]byte, 2)
		ByteOrder.PutUint16(result, uint16(data))
	case int16:
		result = make([]byte, 2)
		ByteOrder.PutUint16(result, uint16(data))
	case uint32:
		result = make([]byte, 4)
		ByteOrder.PutUint32(result, uint32(data))
	case int32:
		result = make([]byte, 4)
		ByteOrder.PutUint32(result, uint32(data))
	case uint64:
		result = make([]byte, 8)
		ByteOrder.PutUint64(result, uint64(data))
	case int64:
		result = make([]byte, 8)
		ByteOrder.PutUint64(result, uint64(data))
	case float32:
		result = make([]byte, 4)
		bits := math.Float32bits(float32(data))
		ByteOrder.PutUint32(result, bits)
	case float64:
		result = make([]byte, 8)
		bits := math.Float64bits(float64(data))
		ByteOrder.PutUint64(result, bits)
	default:
		e = errors.New("This type is not supported " + fmt.Sprint(dest) + "(" + fmt.Sprint(reflect.TypeOf(dest).Kind()) + ")")
		return
	}
	return
}
