package tcp

import (
	"hp-client-golang/Protol"
	"hp-client-golang/hpMessage"
	"net"
)

type LocalProxyUdpHandler struct {
	HpClientHandler *HpClientHandler
	RemoteChannelId string
	Active          bool
}

// ChannelActive 连接激活时，发送注册信息给云端
func (l *LocalProxyUdpHandler) ChannelActive(conn net.Conn) {
	l.Active = true
	l.HpClientHandler.Add(l.RemoteChannelId, conn)
}

func (l *LocalProxyUdpHandler) ChannelRead(conn net.Conn, data interface{}) {
	bytes := data.([]byte)
	message := &hpMessage.HpMessage{
		Type: hpMessage.HpMessage_DATA,
		Data: bytes,
		MetaData: &hpMessage.HpMessage_MetaData{
			Type:      hpMessage.HpMessage_UDP,
			ChannelId: l.RemoteChannelId,
		},
	}
	_, err := l.HpClientHandler.Conn.Write(Protol.Encode(message))
	if err != nil {
		l.HpClientHandler.CallMsg("内网发送远端错误：" + err.Error())
	}
}

func (l *LocalProxyUdpHandler) ChannelInactive(conn net.Conn) {
	l.Active = false
	l.HpClientHandler.Close(l.RemoteChannelId)
}
