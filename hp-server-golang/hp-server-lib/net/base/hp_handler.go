package base

import (
	"github.com/xtaci/smux"
)

// Handler 抽象接口
type HpHandler interface {
	// ChannelActive 连接激活
	ChannelActive(stream *smux.Stream, session *smux.Session)
	// ChannelRead 连接有数据时
	ChannelRead(stream *smux.Stream, data interface{}, session *smux.Session) error
	// ChannelInactive 连接断开
	ChannelInactive(stream *smux.Stream, session *smux.Session)
}
