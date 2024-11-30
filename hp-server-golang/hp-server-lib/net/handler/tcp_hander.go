package handler

import (
	"bufio"
	"io"
	"net"
)

type TcpClientHandler struct {
}

// ChannelActive 连接激活时，发送注册信息给云端
func (h *TcpClientHandler) ChannelActive(conn net.Conn) {

}

func (h *TcpClientHandler) ChannelRead(conn net.Conn, data interface{}) {
	//message := data.(*cmdMessage.CmdMessage)

}

func (h *TcpClientHandler) ChannelInactive(conn net.Conn) {

}

func (h *TcpClientHandler) Decode(reader *bufio.Reader) (error, interface{}) {
	if reader.Buffered() > 0 {
		data := make([]byte, reader.Buffered())
		_, err := io.ReadFull(reader, data)
		return err, data
	}
	return nil, nil
}
