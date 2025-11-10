package handler

import (
	"hp-lib/bean"
	hpMessage "hp-lib/message"
	"hp-lib/protol"
	"net"
)

type LocalSocksProxyHandler struct {
	HpClientHandler *HpClientHandler
	WToN            *bean.WtoN
	Active          bool
}

// ChannelActive 连接激活时，发送注册信息给云端
func (l *LocalSocksProxyHandler) ChannelActive(conn net.Conn) {
	l.Active = true
	l.WToN.N = conn
}

func (l *LocalSocksProxyHandler) ChannelRead(conn net.Conn, data interface{}) {
	bytes := data.([]byte)
	message := &hpMessage.HpMessage{
		Type: hpMessage.HpMessage_DATA,
		Data: bytes,
		MetaData: &hpMessage.HpMessage_MetaData{
			ChannelId: l.WToN.ChannelId,
		},
	}
	err := l.HpClientHandler.writeOutData(l.WToN.W, protol.Encode(message))
	if err != nil {
		l.HpClientHandler.Close(l.WToN.ChannelId)
	}
}

func (l *LocalSocksProxyHandler) ChannelInactive(conn net.Conn) {
	l.Active = false
	l.HpClientHandler.Close(l.WToN.ChannelId)
}
