package handler

import (
	"net"
)

type UdpClientHandler struct {
}

func (h *UdpClientHandler) ChannelRead(conn net.Conn, addr *net.UDPAddr, data interface{}) {
	//message := data.(*cmdMessage.CmdMessage)

}
