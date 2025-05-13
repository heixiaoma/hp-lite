package server

import (
	"errors"
	"github.com/xtaci/smux"
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
func (h *HPClientHandler) ChannelActive(stream *smux.Stream, session *smux.Session) {
	log.Printf("HP传输打开流：%d,%d", stream.ID(), session.NumStreams())

}

func (h *HPClientHandler) ChannelRead(stream *smux.Stream, data interface{}, session *smux.Session) error {
	message := data.(*hpMessage.HpMessage)
	if message == nil {
		log.Printf("消息类型:解码异常|ip:%s", stream.RemoteAddr().String())
		stream.Close()
		return errors.New("HP消息类型:解码异常")
	}
	log.Printf("流ID:%d|HP消息类型:%s|IP:%s", stream.ID(), message.Type.String(), stream.RemoteAddr())
	switch message.Type {
	case hpMessage.HpMessage_REGISTER:
		{
			h.Register(message, session)
		}
	}
	return nil
}

func (h *HPClientHandler) ChannelInactive(stream *smux.Stream, session *smux.Session) {
	if stream != nil {
		log.Printf("HP传输打开流：%d,%d", stream.ID(), session.NumStreams())
	}
}
