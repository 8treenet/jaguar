package plugins

import (
	"bytes"
	"fmt"
	"math/rand"

	"github.com/8treenet/jaguar"
)

func NewSession() *Session {
	return new(Session)
}

type Session struct {
	userId   uint32
	userName string
	Conn     *jaguar.TcpConn `inject:""`
}

func (s *Session) Auth(token string) bool {
	//
	// mock token验证
	//
	s.userName = "user_" + token
	s.userId = uint32(10000 + rand.Intn(80000))

	if GetSession(s.userId) != nil {
		// 已登录
		return false
	}
	SetSession(s.userId, s)
	return true
}

func (s *Session) User() (uname string, uid uint32) {
	uname = s.userName
	uid = s.userId
	return
}

func (s *Session) CloseEvent() {
	fmt.Println("CloseEvent", s.userId)
	RemoveSession(s.userId)
	return
}

func (s *Session) Recover(err error, stack string) {
	fmt.Println("Recover", err, stack)
}

func (s *Session) Reader(protocolId uint16, buffer *bytes.Buffer) *bytes.Buffer {
	fmt.Println("Reader", protocolId)
	return buffer
}

func (s *Session) Writer(protocolId uint16, buffer *bytes.Buffer) *bytes.Buffer {
	fmt.Println("Writer", protocolId)
	return buffer
}

func (s *Session) Request(protocolId uint16, handle interface{}) {
	fmt.Println("Request", protocolId, handle)
}

func (s *Session) Respone(protocolId uint16, handle interface{}) {
	fmt.Println("Respone", protocolId, handle)
}

func (s *Session) Push(protocolId uint16, handle interface{}) {
	fmt.Println("Push", protocolId, handle)
}
