package jaguar

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
)

type ReqHandle interface {
	ReadStream(...interface{}) error
	ReadStreamByString(int, *string) error
	WriteStream(...interface{})
	Respone() error
	ReadStreamBytes() []byte
}

func newReqHandle(protocol uint16, conn *TcpConn, inBuffer *bytes.Buffer) *reqHandle {
	h := new(reqHandle)
	h.protocol = protocol
	h.inBuffer = inBuffer
	h.conn = conn
	return h
}

type reqHandle struct {
	conn       *TcpConn
	inBuffer   *bytes.Buffer
	outBuffer  *bytes.Buffer
	writeError error
	protocol   uint16
}

// ReadStream
func (rh *reqHandle) ReadStream(values ...interface{}) error {
	for index := 0; index < len(values); index++ {
		if err := toType(values[index]); err != nil {
			return err
		}
		if err := binary.Read(rh.inBuffer, _opt.ByteOrder, values[index]); err != nil {
			return err
		}
	}
	return nil
}

func (rh *reqHandle) ReadStreamBytes() []byte {
	return rh.inBuffer.Bytes()
}

func (rh *reqHandle) ReadStreamByString(byteLen int, value *string) error {
	data := rh.inBuffer.Next(byteLen)
	*value = string(data)
	if len(data) == byteLen {
		return nil
	}
	return errors.New("Unknown error")
}

func (rh *reqHandle) WriteStream(values ...interface{}) {
	if rh.outBuffer == nil {
		rh.outBuffer = new(bytes.Buffer)
		proBuf, _ := toBytes(rh.protocol)
		rh.outBuffer.Write(proBuf)
	}
	if rh.writeError != nil {
		return
	}
	for index := 0; index < len(values); index++ {
		data, err := toBytes(values[index])
		if err != nil {
			rh.writeError = err
			break
		}
		_, err = rh.outBuffer.Write(data)
		if err != nil {
			rh.writeError = err
			break
		}
	}
}

func (rh *reqHandle) Respone() error {
	if rh.writeError != nil {
		return rh.writeError
	}
	rh.conn.respone(rh.outBuffer)
	return nil
}

func (rh *reqHandle) di(obj interface{}) {
	structFields(obj, func(sf reflect.StructField, value reflect.Value) {
		if sf.Type.Kind() != reflect.Interface {
			return
		}
		inject, ok := sf.Tag.Lookup("inject")
		if !ok {
			return
		}
		if inject == "req_handle" {
			value.Set(reflect.ValueOf(rh))
			return
		}
		obj := rh.conn.attach.Interface(value.String())
		if obj != nil {
			value.Set(reflect.ValueOf(obj))
		}
	})
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
		_opt.ByteOrder.PutUint16(result, uint16(data))
	case int16:
		result = make([]byte, 2)
		_opt.ByteOrder.PutUint16(result, uint16(data))
	case uint32:
		result = make([]byte, 4)
		_opt.ByteOrder.PutUint32(result, uint32(data))
	case int32:
		result = make([]byte, 4)
		_opt.ByteOrder.PutUint32(result, uint32(data))
	case uint64:
		result = make([]byte, 8)
		_opt.ByteOrder.PutUint64(result, uint64(data))
	case int64:
		result = make([]byte, 8)
		_opt.ByteOrder.PutUint64(result, uint64(data))
	case float32:
		result = make([]byte, 4)
		bits := math.Float32bits(float32(data))
		_opt.ByteOrder.PutUint32(result, bits)
	case float64:
		result = make([]byte, 8)
		bits := math.Float64bits(float64(data))
		_opt.ByteOrder.PutUint64(result, bits)
	default:
		e = errors.New("This type is not supported " + fmt.Sprint(dest) + "(" + fmt.Sprint(reflect.TypeOf(dest).Kind()) + ")")
		return
	}
	return
}

func toType(dest interface{}) (e error) {
	switch dest.(type) {
	case *[]byte:
		if reflect.ValueOf(dest).Elem().Len() == 0 {
			e = errors.New("ReadStream sbyte length is 0")
		}
	case *int8:
	case *uint8:
	case *uint16:
	case *int16:
	case *uint32:
	case *int32:
	case *uint64:
	case *int64:
	case *float32:
	case *float64:
	default:
		e = errors.New("ReadStream type is not supported " + fmt.Sprint(reflect.TypeOf(dest).Kind()))
		return
	}
	return nil
}
