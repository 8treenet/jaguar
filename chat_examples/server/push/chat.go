package push

import (
	"github.com/8treenet/jaguar"
)

func NewChat(sender, content string, row uint32) *chat {
	return &chat{sender: sender, content: content, row: row}
}

type chat struct {
	jaguar.PushHandle `inject:"push_handle"`
	sender            string
	content           string
	row               uint32
}

// ProtocolId - 必须实现
func (c *chat) ProtocolId() uint16 {
	//推送数据包的协议id
	//jaguar.TcpConn 插件会调用 chat.ProtocolId()
	return 300
}

// Encode() - 必须实现
func (c *chat) Encode() {
	//jaguar.TcpConn 插件会调用 chat.Encode()
	c.WriteStream(c.row)
	c.WriteStream(uint8(len(c.sender)), c.sender)
	c.WriteStream(uint16(len(c.content)), c.content)
}
