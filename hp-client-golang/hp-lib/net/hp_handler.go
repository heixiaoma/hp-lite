package net

type HpHandler interface {
	// ChannelActive 连接激活
	ChannelActive(conn *MuxSession)
	// ChannelRead 连接有数据时
	ChannelRead(stream *MuxStream, data interface{})
	// ChannelInactive 连接断开
	ChannelInactive(stream *MuxStream)
}
