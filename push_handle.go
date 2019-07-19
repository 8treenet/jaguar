package jaguar

import (
	"bytes"
	"reflect"
)

type PushHandle interface {
	Buffer() *bytes.Buffer
	WriteStream(values ...interface{})
}

func newPushHandle(protocol uint16) *pushHandle {
	push := new(pushHandle)
	push.outBuffer = new(bytes.Buffer)
	proBuf, _ := toBytes(protocol)
	push.outBuffer.Write(proBuf)
	return push
}

type pushHandle struct {
	outBuffer  *bytes.Buffer
	writeError error
}

func (ph *pushHandle) Buffer() *bytes.Buffer {
	return ph.outBuffer
}

func (ph *pushHandle) WriteStream(values ...interface{}) {
	if ph.writeError != nil {
		return
	}
	for index := 0; index < len(values); index++ {
		data, err := toBytes(values[index])
		if err != nil {
			ph.writeError = err
			break
		}
		_, err = ph.outBuffer.Write(data)
		if err != nil {
			ph.writeError = err
			break
		}
	}
}

func (ph *pushHandle) di(obj interface{}) {
	structFields(obj, func(sf reflect.StructField, value reflect.Value) {
		if sf.Type.Kind() != reflect.Interface {
			return
		}
		inject, ok := sf.Tag.Lookup("inject")
		if !ok {
			return
		}
		if inject == "push_handle" {
			value.Set(reflect.ValueOf(ph))
			return
		}
	})
}
