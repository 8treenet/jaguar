package request

import (
	"fmt"

	"github.com/8treenet/jaguar"
	"github.com/8treenet/jaguar/chat_examples/server/plugins"
	"github.com/8treenet/jaguar/chat_examples/server/push"
)

func init() {
	// 注册请求处理器 对应请求数据包的协议id101
	// 请求处理器chat必须实现 Execute方法
	jaguar.AddRequest(101, new(chat))
}

type chat struct {
	jaguar.ReqHandle `inject:"req_handle"`
	//通过依赖注入获取插件-实体方式
	Session *plugins.Session `inject:""`
	//使用 jaguar.TcpConn 插件, 该插件负责连接相关
	Conn jaguar.TcpConn `inject:"tcp_conn"`
}

// Execute - 必须实现
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

	//太烦了， 超过100条主动断开。
	if row > 10 {
		fmt.Println("More than 10 active disconnections.")
		c.Conn.Close()
	}
}
