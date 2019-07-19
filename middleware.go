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

// Closed - Close registration for callbacks
func (mw *Middleware) Closed(f func()) {
	mw.closed = append(mw.closed, f)
}

// Recover - Request a panic in the code
func (mw *Middleware) Recover(f func(error, string)) {
	mw.recover = append(mw.recover, f)
}

// Reader - Read data
func (mw *Middleware) Reader(f func(uint16, *bytes.Buffer) *bytes.Buffer) {
	mw.reader = append(mw.reader, f)
}

// Writer - Write data
func (mw *Middleware) Writer(f func(uint16, *bytes.Buffer) *bytes.Buffer) {
	mw.writer = append(mw.writer, f)
}

// Respone - Callback when requesting a reply
func (mw *Middleware) Respone(f func(uint16, interface{})) {
	mw.respone = append(mw.respone, f)
}

// Request - Callback at request
func (mw *Middleware) Request(f func(uint16, interface{})) {
	mw.request = append(mw.request, f)
}

// Push - Callbacks on push
func (mw *Middleware) Push(f func(uint16, interface{})) {
	mw.push = append(mw.push, f)
}
