package net

import (
	"github.com/hashicorp/yamux"
)

type HpHandler interface {
	// ChannelActive 连接激活
	ChannelActive(session *yamux.Session)
	// ChannelRead 连接有数据时
	ChannelRead(stream *yamux.Stream, data interface{})
	// ChannelInactive 连接断开
	ChannelInactive(stream *yamux.Stream)
}
