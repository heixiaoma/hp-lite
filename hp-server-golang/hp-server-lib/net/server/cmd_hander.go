package server

import (
	"bufio"
	cmdMessage "hp-server-lib/message"
	"hp-server-lib/protol"
	"hp-server-lib/service"
	"log"
	"net"
)

type CmdClientHandler struct {
	*service.CmdService
}

func NewCmdHandler() *CmdClientHandler {
	return &CmdClientHandler{
		&service.CmdService{},
	}
}

// ChannelActive 连接激活时，发送注册信息给云端
func (h *CmdClientHandler) ChannelActive(conn net.Conn) {
	println("-->ChannelActive")
}

func (h *CmdClientHandler) ChannelRead(conn net.Conn, data interface{}) {
	message := data.(*cmdMessage.CmdMessage)
	log.Printf("消息类型:%s,消息版本:%s,ip:%s", message.Type.String(), message.Version, conn.RemoteAddr().String())
	switch message.Type {
	case cmdMessage.CmdMessage_CONNECT:
		{
			h.Connect(conn, message)
		}
	case cmdMessage.CmdMessage_TIPS:
		{

		}
	case cmdMessage.CmdMessage_DISCONNECT:
		{
			h.Clear(conn)
		}
	}
}

func (h *CmdClientHandler) ChannelInactive(conn net.Conn) {
	println("-->ChannelInactive")
	h.Clear(conn)
}

func (h *CmdClientHandler) Decode(reader *bufio.Reader) (interface{}, error) {
	decode, err := protol.CmdDecode(reader)
	return decode, err
}
