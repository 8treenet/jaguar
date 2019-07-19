package jaguar

import "bytes"

type Middleware struct {
	closed  []func()
	recover []func(error, string)
	request []func(*bytes.Buffer) *bytes.Buffer
	push    []func(*bytes.Buffer) *bytes.Buffer
	respone []func(*bytes.Buffer) *bytes.Buffer
}

func (mw *Middleware) Closed(f func()) {
	mw.closed = append(mw.closed, f)
}

func (mw *Middleware) Recover(f func(error, string)) {
	mw.recover = append(mw.recover, f)
}
func (mw *Middleware) Request(f func(*bytes.Buffer) *bytes.Buffer) {
	mw.request = append(mw.request, f)
}
func (mw *Middleware) Push(f func(*bytes.Buffer) *bytes.Buffer) {
	mw.push = append(mw.push, f)
}

func (mw *Middleware) Respone(f func(*bytes.Buffer) *bytes.Buffer) {
	mw.respone = append(mw.respone, f)
}
