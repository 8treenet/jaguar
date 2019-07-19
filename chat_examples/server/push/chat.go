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

// ProtocolId
func (c *chat) ProtocolId() uint16 {
	return 300
}

// Encode()
func (c *chat) Encode() {
	c.WriteStream(c.row)
	c.WriteStream(uint8(len(c.sender)), c.sender)
	c.WriteStream(uint16(len(c.content)), c.content)
}
