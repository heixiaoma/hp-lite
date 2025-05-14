package base

import (
	"net"
)

// Handler 抽象接口
type UdpHandler interface {
	// ChannelRead 连接有数据时
	ChannelActive(udpConn *net.UDPConn)
	ChannelRead(udpConn *net.UDPConn, data interface{})
	ChannelInactive(udpConn *net.UDPConn)
}
