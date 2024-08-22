package net

import (
	"github.com/quic-go/quic-go"
)

type QuicHandler interface {
	// ChannelActive 连接激活
	ChannelActive(conn quic.Connection)
	// ChannelRead 连接有数据时
	ChannelRead(stream quic.Stream, data interface{})
	// ChannelInactive 连接断开
	ChannelInactive(stream quic.Stream)
}
