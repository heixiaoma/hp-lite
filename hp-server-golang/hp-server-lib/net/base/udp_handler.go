package base

import (
	"net"
)

// Handler 抽象接口
type UdpHandler interface {
	// ChannelRead 连接有数据时
	ChannelRead(conn net.Conn, addr *net.UDPAddr, data interface{})
}
