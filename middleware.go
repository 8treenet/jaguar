package jaguar

import "bytes"

type Middleware struct {
	closed  []func()
	recover []func(error, string)
	reader  []func(uint16, *bytes.Buffer) *bytes.Buffer
	writer  []func(uint16, *bytes.Buffer) *bytes.Buffer
	request []func(uint16, interface{})
	push    []func(uint16, interface{})
	respone []func(uint16, interface{})
}

func (mw *Middleware) Closed(f func()) {
	mw.closed = append(mw.closed, f)
}

func (mw *Middleware) Recover(f func(error, string)) {
	mw.recover = append(mw.recover, f)
}
func (mw *Middleware) Reader(f func(uint16, *bytes.Buffer) *bytes.Buffer) {
	mw.reader = append(mw.reader, f)
}
func (mw *Middleware) Writer(f func(uint16, *bytes.Buffer) *bytes.Buffer) {
	mw.writer = append(mw.writer, f)
}

func (mw *Middleware) Respone(f func(uint16, interface{})) {
	mw.respone = append(mw.respone, f)
}

func (mw *Middleware) Request(f func(uint16, interface{})) {
	mw.request = append(mw.request, f)
}

func (mw *Middleware) Push(f func(uint16, interface{})) {
	mw.push = append(mw.push, f)
}
