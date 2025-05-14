package base

// Handler 抽象接口
type HpHandler interface {
	// ChannelActive 连接激活
	ChannelActive(stream *MuxStream, conn *MuxSession)
	// ChannelRead 连接有数据时
	ChannelRead(stream *MuxStream, data interface{}, conn *MuxSession) error
	// ChannelInactive 连接断开
	ChannelInactive(stream *MuxStream, conn *MuxSession)
}
