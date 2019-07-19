package request

import (
	"fmt"

	"github.com/8treenet/jaguar"
	"github.com/8treenet/jaguar/chat_examples/server/plugins"
)

func init() {
	jaguar.AddRequest(100, new(login))
}

type login struct {
	jaguar.ReqHandle `inject:"req_handle"`
	Session          *plugins.Session `inject:""`
}

// Execute
func (l *login) Execute() {
	//1. ReqHandle.ReadStream 读字节流
	//2. ReqHandle.WriteStream 写字节流.
	//3. 如果使用protocol和json 支持自行定义，使用 ReqHandle.ReadStreamBytes 读取全部 []byte
	var bytesLen uint8
	var token string
	l.ReadStream(&bytesLen)
	l.ReadStreamByString(int(bytesLen), &token)
	auth := l.Session.Auth(token)
	if auth {
		//成功
		uname, uid := l.Session.User()
		l.WriteStream(uint8(1))
		l.WriteStream(uid)

		unameSize := uint8(len(uname))
		l.WriteStream(unameSize, uname)
		fmt.Println(fmt.Sprintf("Access success user_id:%d, user_name: %s", uid, uname))
	} else {
		//失败
		l.WriteStream(uint8(2))
	}

	l.Respone()
}
