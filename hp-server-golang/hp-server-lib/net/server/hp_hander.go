package server

import (
	"errors"
	hpMessage "hp-server-lib/message"
	net2 "hp-server-lib/net/base"
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
func (h *HPClientHandler) ChannelActive(stream *net2.MuxStream, conn *net2.MuxSession) {
	if stream.IsTcp {
		log.Printf("HP传输打开流：%d", stream.TcpStream.ID())
	} else {
		log.Printf("HP传输打开流：%d", stream.QuicStream.StreamID())
	}
}

func (h *HPClientHandler) ChannelRead(stream *net2.MuxStream, data interface{}, conn *net2.MuxSession) error {
	message := data.(*hpMessage.HpMessage)
	if message == nil {
		if stream.IsTcp {
			log.Printf("消息类型:解码异常|ip:%s", conn.TcpSession.RemoteAddr().String())
			stream.TcpStream.Close()
		} else {
			log.Printf("消息类型:解码异常|ip:%s", conn.QuicSession.RemoteAddr().String())
			stream.QuicStream.Close()
		}
		return errors.New("HP消息类型:解码异常")
	}

	if stream.IsTcp {
		log.Printf("流ID:%d|HP消息类型:%s|IP:%s", stream.TcpStream.ID(), message.Type.String(), conn.TcpSession.RemoteAddr())
	} else {
		log.Printf("流ID:%d|HP消息类型:%s|IP:%s", stream.QuicStream.StreamID(), message.Type.String(), conn.QuicSession.RemoteAddr())
	}
	switch message.Type {
	case hpMessage.HpMessage_REGISTER:
		{
			h.Register(message, conn)
		}
	}
	return nil
}

func (h *HPClientHandler) ChannelInactive(stream *net2.MuxStream, conn *net2.MuxSession) {
	if stream != nil {
		if stream.IsTcp && stream.TcpStream != nil {
			log.Printf("HP传输断开流：%d", stream.TcpStream.ID())
		} else if stream.QuicStream != nil {
			log.Printf("HP传输断开流：%d", stream.QuicStream.StreamID())
		}
	}
}
