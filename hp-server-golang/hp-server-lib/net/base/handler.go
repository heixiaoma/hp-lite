package base

import (
	"bufio"
	"net"
)

// Handler 抽象接口
type Handler interface {
	// ChannelActive 连接激活
	ChannelActive(conn net.Conn)
	// ChannelRead 连接有数据时
	ChannelRead(conn net.Conn, data interface{}) error
	// ChannelInactive 连接断开
	ChannelInactive(conn net.Conn)
	//消息解码器
	Decode(reader *bufio.Reader) (interface{}, error)
}
