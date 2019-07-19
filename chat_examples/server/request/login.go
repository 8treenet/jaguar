package request

import (
	"fmt"

	"github.com/8treenet/jaguar"
)

func init() {
	// 注册请求处理器 对应请求数据包的协议id101
	// 请求处理器login必须实现 Execute方法
	jaguar.AddRequest(100, new(login))
}

type login struct {
	//继承 jaguar.ReqHandle 接口
	jaguar.ReqHandle `inject:"req_handle"`
	//通过依赖注入获取插件-接口方式
	Session implExample `inject:"impl_example"`
}

// 定义一个接口 和 plugins.Session负责实现
type implExample interface {
	Auth(string) bool
	User() (string, uint32)
}

// Execute - 必须实现
func (l *login) Execute() {
	//如果使用protocol和json 可自行定义，使用 ReqHandle.ReadStreamBytes 读取全部 []byte后解析
	var bytesLen uint8
	var token string
	//ReqHandle.ReadStream 读字节流
	l.ReadStream(&bytesLen)
	l.ReadStreamByString(int(bytesLen), &token)
	auth := l.Session.Auth(token)
	if auth {
		//成功
		uname, uid := l.Session.User()
		//ReqHandle.WriteStream 写字节流.
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
