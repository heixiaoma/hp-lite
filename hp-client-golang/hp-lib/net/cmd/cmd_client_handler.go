package cmd

import (
	"hp-lib/bean"
	cmdMessage "hp-lib/message"
	"hp-lib/net/hp"
	"hp-lib/protol"
	"hp-lib/util"
	"net"
	"os"
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
		Data:    util.SysInfo(),
	}
	conn.Write(protol.CmdEncode(message))
}

func (h *CmdClientHandler) ChannelRead(conn net.Conn, data interface{}) {
	message := data.(*cmdMessage.CmdMessage)
	switch message.Type {
	case cmdMessage.CmdMessage_DISCONNECT:
		h.CmdClient.CallMsg("服务器要求你关闭：" + message.GetData())
		h.CmdClient.Close()
		os.Exit(-1)
		break
	case cmdMessage.CmdMessage_TIPS:
		h.CmdClient.CallMsg(message.Data)
		conn.Write(protol.CmdEncode(&cmdMessage.CmdMessage{Version: version, Key: h.Key, Type: cmdMessage.CmdMessage_TIPS, Data: util.SysInfo()}))
		break

	case cmdMessage.CmdMessage_LOCAL_INNER_WEAR:
		h.CmdClient.CallMsg("正在检查本地映射配置关系")
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
	hp.RefreshRouter(wear, h.CmdClient.CallMsg)
}
