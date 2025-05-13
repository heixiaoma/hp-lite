package net

import (
	"github.com/xtaci/smux"
)

type HpHandler interface {
	// ChannelActive 连接激活
	ChannelActive(session *smux.Session)
	// ChannelRead 连接有数据时
	ChannelRead(stream *smux.Stream, data interface{})
	// ChannelInactive 连接断开
	ChannelInactive(stream *smux.Stream)
}
