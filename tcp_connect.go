package jaguar

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"reflect"
	"runtime/debug"
	"time"
)

func init() {
	routeMap = make(map[uint16]interface{})
}

var routeMap map[uint16]interface{}

// AddRequest
func AddRequest(protocolId uint16, req ReqHandle) {
	type execute interface {
		Execute()
	}
	_, ok := req.(execute)
	if !ok {
		panic(fmt.Sprintf("id : %d", protocolId) + " No Execute implementation method")
	}
	routeMap[protocolId] = req
}

// newConn
func newConn(conn net.Conn, hook *Middleware) *TcpConn {
	ts := new(TcpConn)
	ts.Conn = conn
	ts.writeBuffer = make(chan []byte, 4096)
	ts.attach = NewJMap()
	ts.Attach(ts)
	ts.hook = hook
	return ts
}

type TcpConn struct {
	net.Conn
	writeBuffer chan []byte
	attach      *JMap
	hook        *Middleware
}

// start
func (tc *TcpConn) start() {
	go tc.read()
	go tc.write()
}

// write
func (tc *TcpConn) write() {
	for {
		data := <-tc.writeBuffer
		head := len(data)
		if head == 0 {
			return
		}

		headData := tc.uintToBytes(uint32(head))
		tc.SetWriteDeadline(time.Now().Add(time.Second * 15))
		if _, err := tc.Write(headData); err != nil {
			return
		}
		if _, err := tc.Write(data); err != nil {
			return
		}
	}
}

// packetSize
func (tc *TcpConn) packetSize(head []byte) uint32 {
	buffer := bytes.NewBuffer(head)
	var number uint32
	binary.Read(buffer, binary.BigEndian, &number)
	return number
}

// break
func (tc *TcpConn) readBreak() {
	for _, f := range tc.hook.closed {
		f()
	}
}

//Close 关闭
func (tc *TcpConn) Close() {
	tc.writeBuffer <- make([]byte, 0)
	tc.Conn.Close()
}

func (tc *TcpConn) send(data []byte) {
	tc.writeBuffer <- data
}

type Encode interface {
	Encode()
	ProtocolId() uint16
}

func (tc *TcpConn) Push(pushHandle Encode) error {
	ph := newPushHandle(pushHandle.ProtocolId())
	ph.di(pushHandle)
	pushHandle.Encode()
	if ph.outBuffer == nil || ph.outBuffer.Len() == 0 {
		return errors.New("PushHandle doesn't writeStream")
	}
	if ph.writeError == nil {
		tc.send(ph.outBuffer.Bytes())
	}
	return ph.writeError
}

func (tc *TcpConn) respone(buffer *bytes.Buffer) {
	for _, f := range tc.hook.respone {
		buffer = f(buffer)
	}
	sendData := buffer.Bytes()
	tc.send(sendData)
}

// route
func (tc *TcpConn) route(buffer *bytes.Buffer) {
	var id uint16 = 0
	defer func() {
		if perr := recover(); perr != nil {
			e := errors.New(fmt.Sprint(perr))
			stack := string(debug.Stack())
			for _, f := range tc.hook.recover {
				f(e, stack)
			}
		}
	}()

	for _, f := range tc.hook.request {
		buffer = f(buffer)
	}
	binary.Read(buffer, binary.BigEndian, &id)
	req, ok := routeMap[id]
	if !ok {
		return
	}
	t := reflect.TypeOf(req)
	if t == nil {
		tc.Close()
		return
	}
	newReq := reflect.New(t.Elem())
	handle := newReqHandle(id, tc, buffer)
	handle.di(newReq.Interface())
	tc.di(newReq.Interface())
	newReq.MethodByName("Execute").Call(nil)
}

func (tc *TcpConn) uintToBytes(number uint32) []byte {
	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, number)
	return buffer
}

// Attach
func (tc *TcpConn) Attach(obj interface{}) {
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() != reflect.Ptr {
		panic("Use a pointer object, The " + fmt.Sprint(obj))
	}
	if tc.attach.Exist(obj) {
		panic("Object has been registered, The " + fmt.Sprint(obj))
	}
	tc.attach.Set(reflect.TypeOf(obj).String(), obj)
}

func (tc *TcpConn) di(child interface{}) {
	structFields(child, func(sf reflect.StructField, value reflect.Value) {
		_, inject := sf.Tag.Lookup("inject")
		if !inject {
			return
		}
		obj := tc.attach.Interface(value.Type().String())
		if obj != nil {
			value.Set(reflect.ValueOf(obj))
		}
	})
}

func (tc *TcpConn) attachDi() {
	for _, obj := range tc.attach._map {
		tc.di(obj)
	}
}

func (tc *TcpConn) Read(b []byte) (n int, err error) {
	panic("TcpConn.Read")
}

// read
func (tc *TcpConn) read() {
	action := 1
	bodyLen := uint32(0)
	packet := new(bytes.Buffer)
	for {
		tc.SetReadDeadline(time.Now().Add(_opt.IdleCheckFrequency))
		packet.Reset()
		var data []byte
		var trialLen int
		if action == 1 {
			trialLen = int(_opt.PacketHeadLen)
		} else {
			trialLen = int(bodyLen)
		}

		for trialLen > 0 {
			data = make([]byte, trialLen)
			len, derr := tc.Conn.Read(data)
			if len == 0 || derr != nil {
				tc.Conn.Close()
				tc.readBreak()
				return
			}

			packet.Write(data[:len])
			trialLen = trialLen - len
		}

		if action == 1 {
			action = 2
			bodyLen = tc.packetSize(packet.Bytes())
			if bodyLen > _opt.PacketMaximum {
				tc.Conn.Close()
				tc.readBreak()
				return
			}
		} else {
			action = 1
			packetHandle := new(bytes.Buffer)
			packetHandle.Write(packet.Bytes())
			go tc.route(packetHandle)
		}
	}
}
