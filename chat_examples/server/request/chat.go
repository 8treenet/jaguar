package request

import (
	"github.com/8treenet/jaguar"
	"github.com/8treenet/jaguar/chat_examples/server/plugins"
	"github.com/8treenet/jaguar/chat_examples/server/push"
)

func init() {
	jaguar.AddRequest(101, new(chat))
}

type chat struct {
	jaguar.ReqHandle `inject:"req_handle"`
	Session          *plugins.Session `inject:""`
	Conn             *jaguar.TcpConn  `inject:""`
}

func (c *chat) Execute() {
	var (
		row        uint32
		contentLen uint16
		content    string
	)
	c.ReadStream(&row, &contentLen)
	c.ReadStreamByString(int(contentLen), &content)
	//回执成功
	c.WriteStream(uint8(1))
	c.Respone()
	name, _ := c.Session.User()
	allSession := plugins.AllSession()
	for index := 0; index < len(allSession); index++ {
		if c.Session == allSession[index] {
			continue
		}
		packet := push.NewChat(name, content, row)
		allSession[index].Conn.Push(packet)
	}
}
