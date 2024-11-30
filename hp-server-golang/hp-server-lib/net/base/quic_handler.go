package base

import (
	"github.com/quic-go/quic-go"
)

// Handler 抽象接口
type QuicHandler interface {
	// ChannelActive 连接激活
	ChannelActive(conn quic.Stream)
	// ChannelRead 连接有数据时
	ChannelRead(conn quic.Stream, data interface{})
	// ChannelInactive 连接断开
	ChannelInactive(conn quic.Stream)
}
