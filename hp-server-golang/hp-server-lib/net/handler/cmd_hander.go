package handler

import (
	"bufio"
	"hp-server-lib/protol"
	"net"
)

type CmdClientHandler struct {
}

// ChannelActive 连接激活时，发送注册信息给云端
func (h *CmdClientHandler) ChannelActive(conn net.Conn) {

}

func (h *CmdClientHandler) ChannelRead(conn net.Conn, data interface{}) {
	//message := data.(*cmdMessage.CmdMessage)

}

func (h *CmdClientHandler) ChannelInactive(conn net.Conn) {

}

func (h *CmdClientHandler) Decode(reader *bufio.Reader) (error, interface{}) {
	decode, err := protol.CmdDecode(reader)
	return err, decode
}
