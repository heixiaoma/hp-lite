package handler

import (
	"hp-lib/bean"
	hpMessage "hp-lib/message"
	"hp-lib/protol"
	"log"
	"net"
)

type LocalProxyUdpHandler struct {
	HpClientHandler *HpClientHandler
	WToN            *bean.WtoN
	Active          bool
}

// ChannelActive 连接激活时，发送注册信息给云端
func (l *LocalProxyUdpHandler) ChannelActive(conn net.Conn) {
	l.Active = true
	l.WToN.N = conn
}

func (l *LocalProxyUdpHandler) ChannelRead(conn net.Conn, data interface{}) {
	bytes := data.([]byte)
	message := &hpMessage.HpMessage{
		Type: hpMessage.HpMessage_DATA,
		Data: bytes,
		MetaData: &hpMessage.HpMessage_MetaData{
			Type:      hpMessage.HpMessage_UDP,
			ChannelId: l.WToN.ChannelId,
		},
	}
	log.Printf("发送UDP" + l.WToN.ChannelId)
	err := l.HpClientHandler.writeOutData(l.WToN.W, protol.Encode(message))
	if err != nil {
		l.HpClientHandler.CallMsg("UDP内网发送远端错误：" + err.Error())
		l.HpClientHandler.Close(l.WToN.ChannelId)
	}
}

func (l *LocalProxyUdpHandler) ChannelInactive(conn net.Conn) {
	l.Active = false
	l.HpClientHandler.Close(l.WToN.ChannelId)
}
