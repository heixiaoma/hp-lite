package cmd

import (
	"hp-client-golang/Protol"
	"hp-client-golang/bean"
	"hp-client-golang/cmdMessage"
	"hp-client-golang/tcp"
	"net"
)

type CmdClientHandler struct {
	Key       string
	CmdClient *CmdClient
	Conn      net.Conn
	Active    bool
}

// ChannelActive 连接激活时，发送注册信息给云端
func (h *CmdClientHandler) ChannelActive(conn net.Conn) {
	h.Conn = conn
	h.Active = true
	message := &cmdMessage.CmdMessage{
		Version: version,
		Type:    cmdMessage.CmdMessage_CONNECT,
		Key:     h.Key,
	}
	conn.Write(Protol.CmdEncode(message))
}

func (h *CmdClientHandler) ChannelRead(conn net.Conn, data interface{}) {
	message := data.(*cmdMessage.CmdMessage)
	switch message.Type {
	case cmdMessage.CmdMessage_DISCONNECT:
		h.CmdClient.CallMsg("服务器要求你关闭：" + message.GetData())
		h.CmdClient.Close()
		break
	case cmdMessage.CmdMessage_TIPS:
		h.CmdClient.CallMsg(message.Data)
		conn.Write(Protol.CmdEncode(&cmdMessage.CmdMessage{Version: version, Key: h.Key, Type: cmdMessage.CmdMessage_TIPS}))
		break

	case cmdMessage.CmdMessage_LOCAL_INNER_WEAR:
		h.connected(message)
		break
	default:
		h.CmdClient.CallMsg("未知类型数据：" + message.GetData())
	}
}

func (h *CmdClientHandler) ChannelInactive(conn net.Conn) {
	h.Active = false
	conn.Close()
}

func (h *CmdClientHandler) connected(message *cmdMessage.CmdMessage) {
	//如果是TCP数据包，我们就连接本地的TCP服务器
	wear := bean.NewLocalInnerWear(message.Data)
	tcp.RefreshRouter(wear, h.CmdClient.CallMsg)
}
