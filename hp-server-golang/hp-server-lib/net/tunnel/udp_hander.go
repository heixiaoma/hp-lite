package tunnel

import (
	"github.com/quic-go/quic-go"
	"net"
)

type UdpHandler struct {
	conn      quic.Connection
	stream    quic.Stream
	channelId string
}

func NewUdpHandler(conn quic.Connection) *UdpHandler {
	return &UdpHandler{conn: conn}
}

func (h *UdpHandler) ChannelRead(conn net.Conn, addr *net.UDPAddr, data interface{}) {

	//message := data.(*cmdMessage.CmdMessage)

}
