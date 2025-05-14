package server

import (
	"bufio"
	"errors"
	cmdMessage "hp-server-lib/message"
	"hp-server-lib/protol"
	"hp-server-lib/service"
	"log"
	"net"
	"runtime/debug"
	"time"
)

const (
	// 最大闲置时间
	IdleTimeout = 5 * time.Minute
)

type CmdClientHandler struct {
	*service.CmdService
	IdleTimer *time.Timer
}

func NewCmdHandler() *CmdClientHandler {
	return &CmdClientHandler{
		&service.CmdService{},
		time.NewTimer(IdleTimeout),
	}
}

func (receiver CmdClientHandler) idle(conn net.Conn) {
	go func() {
		for {
			select {
			case <-receiver.IdleTimer.C:
				if conn == nil {
					return
				}
				c := &cmdMessage.CmdMessage{
					Data: "中心节点-心跳数据",
					Type: cmdMessage.CmdMessage_TIPS,
				}
				// 如果 5 分钟没有读写操作，发送心跳包
				_, err := conn.Write(protol.CmdEncode(c))
				if err != nil {
					return
				}
				receiver.IdleTimer.Reset(IdleTimeout)
			}
		}
	}()
}

// ChannelActive 连接激活时，发送注册信息给云端
func (h *CmdClientHandler) ChannelActive(conn net.Conn) {
	log.Printf("CMD指令激活 ip:%s", conn.RemoteAddr().String())
	h.idle(conn)
}

func (h *CmdClientHandler) ChannelRead(conn net.Conn, data interface{}) error {
	defer func() {
		if err := recover(); err != nil {
			// 捕获异常并记录日志
			log.Printf("CMD-ChannelRead: %v\n栈情况: %s", err, string(debug.Stack()))
		}
	}()
	message := data.(*cmdMessage.CmdMessage)
	if message == nil {
		log.Printf("CMD消息类型:解码异常|ip:%s", conn.RemoteAddr().String())
		return errors.New("消息类型异常")
	}
	log.Printf("消息类型:%s|消息版本:%s|ip:%s", message.Type.String(), message.Version, conn.RemoteAddr().String())
	switch message.Type {
	case cmdMessage.CmdMessage_CONNECT:
		{
			h.Connect(conn, message)
		}
	case cmdMessage.CmdMessage_TIPS:
		{
			h.StoreMemInfo(conn, message)
		}
	case cmdMessage.CmdMessage_DISCONNECT:
		{
			h.Clear(conn)
		}
	}
	return nil

}

func (h *CmdClientHandler) ChannelInactive(conn net.Conn) {
	log.Printf("CMD指令断开 ip:%s", conn.RemoteAddr().String())
	h.Clear(conn)
}

func (h *CmdClientHandler) Decode(reader *bufio.Reader) (interface{}, error) {
	decode, err := protol.CmdDecode(reader)
	return decode, err
}
