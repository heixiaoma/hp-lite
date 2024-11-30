package handler

import (
	"bufio"
	"io"
	"net"
)

type TcpClientHandler struct {
}

func NewTcpClientHandler() *TcpClientHandler {
	return &TcpClientHandler{}
}

// ChannelActive 连接激活时，发送注册信息给云端
func (h *TcpClientHandler) ChannelActive(conn net.Conn) {

}

func (h *TcpClientHandler) ChannelRead(conn net.Conn, data interface{}) {
	//message := data.(*cmdMessage.CmdMessage)

}

func (h *TcpClientHandler) ChannelInactive(conn net.Conn) {

}

func (h *TcpClientHandler) Decode(reader *bufio.Reader) (interface{}, error) {
	if reader.Buffered() > 0 {
		data := make([]byte, reader.Buffered())
		_, err := io.ReadFull(reader, data)
		return data, err
	}
	return nil, nil
}
