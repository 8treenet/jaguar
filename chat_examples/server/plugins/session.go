package plugins

import (
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
	RemoveSession(s.userId)
	return
}
