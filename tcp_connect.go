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

type TcpConn interface {
	//Additional plug-in objects, available by relying on injection
	Attach(plugin interface{})
	//Additional plug-in objects, interface sits, which can be obtained by relying on injection
	AttachImpl(impl string, plugin interface{})
	//Actively close the connection
	Close()
	//Remote address connected
	RemoteAddr() net.Addr
	//Push
	Push(pushHandle Encode) error
}

func init() {
	routeMap = make(map[uint16]interface{})
}

var routeMap map[uint16]interface{}

// AddRequest - Join the request processor
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
func newConn(conn net.Conn, hook *Middleware) *tcpConn {
	ts := new(tcpConn)
	ts.Conn = conn
	ts.writeBuffer = make(chan []byte, 4096)
	ts.attach = NewJMap()
	ts.AttachImpl("tcp_conn", ts)
	ts.hook = hook
	return ts
}

type tcpConn struct {
	net.Conn
	writeBuffer chan []byte
	attach      *JMap
	attachImpl  *JMap
	hook        *Middleware
}

// start
func (tc *tcpConn) start() {
	go tc.read()
	go tc.write()
}

// write
func (tc *tcpConn) write() {
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
func (tc *tcpConn) packetSize(head []byte) uint32 {
	buffer := bytes.NewBuffer(head)
	var number uint32
	binary.Read(buffer, binary.BigEndian, &number)
	return number
}

// break
func (tc *tcpConn) readBreak() {
	for _, f := range tc.hook.closed {
		f()
	}
}

//Close 关闭
func (tc *tcpConn) Close() {
	tc.writeBuffer <- make([]byte, 0)
	tc.Conn.Close()
}

func (tc *tcpConn) send(data []byte) {
	if len(data) == 0 {
		return
	}
	tc.writeBuffer <- data
}

type Encode interface {
	Encode()
	ProtocolId() uint16
}

func (tc *tcpConn) Push(pushHandle Encode) error {
	protocolId := pushHandle.ProtocolId()
	ph := newPushHandle(protocolId)
	ph.di(pushHandle)
	for _, f := range tc.hook.push {
		f(protocolId, pushHandle)
	}

	pushHandle.Encode()
	if ph.outBuffer == nil || ph.outBuffer.Len() == 0 {
		return errors.New("PushHandle doesn't writeStream")
	}

	if ph.writeError == nil {
		for _, f := range tc.hook.writer {
			ph.outBuffer = f(protocolId, ph.outBuffer)
		}
		tc.send(ph.outBuffer.Bytes())
	}
	return ph.writeError
}

func (tc *tcpConn) respone(protocolId uint16, buffer *bytes.Buffer) {
	for _, f := range tc.hook.writer {
		buffer = f(protocolId, buffer)
	}
	sendData := buffer.Bytes()
	tc.send(sendData)
}

// route
func (tc *tcpConn) route(buffer *bytes.Buffer) {
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
	binary.Read(buffer, binary.BigEndian, &id)
	for _, f := range tc.hook.reader {
		buffer = f(id, buffer)
	}
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
	for _, f := range tc.hook.request {
		f(id, newReq.Interface())
	}
	newReq.MethodByName("Execute").Call(nil)
}

func (tc *tcpConn) uintToBytes(number uint32) []byte {
	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, number)
	return buffer
}

// Attach
func (tc *tcpConn) Attach(obj interface{}) {
	if obj == nil {
		panic("obj not nil")
	}
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() != reflect.Ptr {
		panic("Use a pointer object, The " + fmt.Sprint(obj))
	}
	if tc.attach.Exist(obj) {
		panic("Object has been registered, The " + fmt.Sprint(obj))
	}
	tc.attach.Set(reflect.TypeOf(obj).String(), obj)
}

// AttachImpl
func (tc *tcpConn) AttachImpl(impl string, obj interface{}) {
	if obj == nil {
		panic("obj not nil")
	}
	tc.attach.Set("inject%_%"+impl, obj)
}

func (tc *tcpConn) di(child interface{}) {
	structFields(child, func(sf reflect.StructField, value reflect.Value) {
		tag, inject := sf.Tag.Lookup("inject")
		if !inject {
			return
		}
		if tag != "" {
			obj := tc.attach.Interface("inject%_%" + tag)
			if obj != nil {
				value.Set(reflect.ValueOf(obj))
			}
			return
		}
		obj := tc.attach.Interface(value.Type().String())
		if obj != nil {
			value.Set(reflect.ValueOf(obj))
		}
	})
}

func (tc *tcpConn) attachDi() {
	for _, obj := range tc.attach._map {
		tc.di(obj)
	}
}

// read
func (tc *tcpConn) read() {
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
