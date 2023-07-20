package tcp

import (
	"hp-client-golang/Protol"
	"hp-client-golang/hpMessage"
	"net"
)

type LocalProxyHandler struct {
	HpClientHandler *HpClientHandler
	RemoteChannelId string
	Active          bool
}

// ChannelActive 连接激活时，发送注册信息给云端
func (l *LocalProxyHandler) ChannelActive(conn net.Conn) {
	l.Active = true
	l.HpClientHandler.Add(l.RemoteChannelId, conn)
}

func (l *LocalProxyHandler) ChannelRead(conn net.Conn, data interface{}) {
	bytes := data.([]byte)
	message := &hpMessage.HpMessage{
		Type: hpMessage.HpMessage_DATA,
		Data: bytes,
		MetaData: &hpMessage.HpMessage_MetaData{
			Type:      hpMessage.HpMessage_TCP,
			ChannelId: l.RemoteChannelId,
		},
	}
	l.HpClientHandler.Conn.Write(Protol.Encode(message))
}

func (l *LocalProxyHandler) ChannelInactive(conn net.Conn) {
	l.Active = false
	l.HpClientHandler.Close(l.RemoteChannelId)
}
