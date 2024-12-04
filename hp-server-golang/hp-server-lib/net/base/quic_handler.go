package base

import (
	"github.com/quic-go/quic-go"
)

// Handler 抽象接口
type QuicHandler interface {
	// ChannelActive 连接激活
	ChannelActive(stream quic.Stream, conn quic.Connection)
	// ChannelRead 连接有数据时
	ChannelRead(stream quic.Stream, data interface{}, conn quic.Connection) error
	// ChannelInactive 连接断开
	ChannelInactive(stream quic.Stream, conn quic.Connection)
}
