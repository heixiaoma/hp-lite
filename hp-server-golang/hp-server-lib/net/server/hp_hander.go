package server

import (
	"errors"
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
	log.Printf("HP传输打开流：%d", stream.StreamID())

}

func (h *HPClientHandler) ChannelRead(stream quic.Stream, data interface{}, conn quic.Connection) error {
	message := data.(*hpMessage.HpMessage)
	if message == nil {
		log.Printf("消息类型:解码异常|ip:%s", conn.RemoteAddr().String())
		stream.Close()
		return errors.New("HP消息类型:解码异常")
	}
	log.Printf("流ID:%d|HP消息类型:%s|IP:%s", stream.StreamID(), message.Type.String(), conn.RemoteAddr())
	switch message.Type {
	case hpMessage.HpMessage_REGISTER:
		{
			h.Register(message, conn)
		}
	}
	return nil
}

func (h *HPClientHandler) ChannelInactive(stream quic.Stream, conn quic.Connection) {
	if stream != nil {
		log.Printf("HP传输断开流：%d", stream.StreamID())
	}
}
