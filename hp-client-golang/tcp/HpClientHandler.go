package tcp

import (
	"encoding/json"
	"hp-client-golang/Protol"
	"hp-client-golang/hpMessage"
	"net"
	"sync"
)

type HpClientHandler struct {
	Key          string
	MessageType  hpMessage.HpMessage_MessageType
	ProxyAddress string
	ProxyPort    int
	CallMsg      func(message string)
	Conn         net.Conn
	Active       bool
}

var ConnGroup = sync.Map{}

// ChannelActive 连接激活时，发送注册信息给云端
func (h *HpClientHandler) ChannelActive(conn net.Conn) {
	h.Conn = conn
	h.Active = true
	message := &hpMessage.HpMessage{
		Type: hpMessage.HpMessage_REGISTER,
		MetaData: &hpMessage.HpMessage_MetaData{
			Key: h.Key,
		},
	}
	message.MetaData.Type = h.MessageType
	conn.Write(Protol.Encode(message))
}

func (h *HpClientHandler) ChannelRead(conn net.Conn, data interface{}) {
	message := data.(*hpMessage.HpMessage)
	switch message.Type {
	case hpMessage.HpMessage_REGISTER_RESULT:
		h.CallMsg(message.MetaData.Reason)
		break
	case hpMessage.HpMessage_CONNECTED:
		h.connected(message)
		break
	case hpMessage.HpMessage_DISCONNECTED:
		h.Close(message.MetaData.ChannelId)
		break
	case hpMessage.HpMessage_DATA:
		h.writeData(message)
		break
	case hpMessage.HpMessage_KEEPALIVE:
		h.CallMsg("服务器端返回心跳数据")
		conn.Write(Protol.Encode(&hpMessage.HpMessage{Type: hpMessage.HpMessage_KEEPALIVE}))
		break
	default:
		marshal, _ := json.Marshal(message.MetaData)
		h.CallMsg("未知类型数据：" + string(marshal))
	}
}

func (h *HpClientHandler) ChannelInactive(conn net.Conn) {
	h.Active = false
}

func (h *HpClientHandler) connected(message *hpMessage.HpMessage) {
	//如果是TCP数据包，我们就连接本地的TCP服务器
	if message.MetaData.Type == hpMessage.HpMessage_TCP {
		NewTcpConnection().ConnectLocal(h.ProxyAddress, h.ProxyPort, &LocalProxyHandler{
			HpClientHandler: h,
			RemoteChannelId: message.MetaData.ChannelId,
		}, h.CallMsg)
	}
	if message.MetaData.Type == hpMessage.HpMessage_UDP {
		NewUdpConnection().Connect(h.ProxyAddress, h.ProxyPort, &LocalProxyUdpHandler{
			HpClientHandler: h,
			RemoteChannelId: message.MetaData.ChannelId,
		}, h.CallMsg)
	}
}
func (h *HpClientHandler) Add(channelId string, conn net.Conn) {
	ConnGroup.Store(channelId, conn)
}

func (h *HpClientHandler) Close(channelId string) {
	load, ok := ConnGroup.Load(channelId)
	if ok {
		conn := load.(net.Conn)
		if conn != nil {
			conn.Close()
			ConnGroup.Delete(channelId)
		}
	}
}

func (h *HpClientHandler) CloseAll() {
	ConnGroup.Range(func(key, value interface{}) bool {
		conn := value.(net.Conn)
		if conn != nil {
			conn.Close()
			ConnGroup.Delete(key)
		}
		return true
	})
}

func (h *HpClientHandler) writeData(message *hpMessage.HpMessage) {
	if message.MetaData.Type == hpMessage.HpMessage_TCP {
		load, ok := ConnGroup.Load(message.MetaData.ChannelId)
		if ok {
			conn := load.(net.Conn)
			if conn != nil {
				conn.Write(message.Data)
			}
		}

	} else if message.MetaData.Type == hpMessage.HpMessage_UDP {
		//todo 需要转发给内网的UDPServer
		load, ok := ConnGroup.Load(message.MetaData.ChannelId)
		if ok {
			conn := load.(net.Conn)
			if conn != nil {
				conn.Write(message.Data)
			}
		}
	}
}
