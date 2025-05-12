package base

import (
	"github.com/hashicorp/yamux"
)

// Handler 抽象接口
type HpHandler interface {
	// ChannelActive 连接激活
	ChannelActive(stream *yamux.Stream, session *yamux.Session)
	// ChannelRead 连接有数据时
	ChannelRead(stream *yamux.Stream, data interface{}, session *yamux.Session) error
	// ChannelInactive 连接断开
	ChannelInactive(stream *yamux.Stream, session *yamux.Session)
}
