package server

import (
	"github.com/quic-go/quic-go"
	hpMessage "hp-server-lib/message"
	"hp-server-lib/service"
	"log"
)

type HPClientHandler struct {
	*service.HpService
}

func NewHPHandler() *HPClientHandler {
	return &HPClientHandler{
		&service.HpService{},
	}
}

// ChannelActive 连接激活时，发送注册信息给云端
func (h *HPClientHandler) ChannelActive(stream quic.Stream, conn quic.Connection) {
	println("HPClientHandler-->ChannelActive")
}

func (h *HPClientHandler) ChannelRead(stream quic.Stream, data interface{}, conn quic.Connection) {
	message := data.(*hpMessage.HpMessage)
	log.Printf("消息类型:%s,ip:%s", message.Type.String(), conn)
	switch message.Type {
	case hpMessage.HpMessage_REGISTER:
		{
			h.Register(message, conn)
		}
	}
}

func (h *HPClientHandler) ChannelInactive(stream quic.Stream, conn quic.Connection) {
	println("HPClientHandler-->ChannelInactive")
}
