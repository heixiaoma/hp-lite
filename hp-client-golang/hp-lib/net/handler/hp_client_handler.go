package handler

import (
	"context"
	"encoding/json"
	"github.com/quic-go/quic-go"
	"golang.org/x/time/rate"
	"hp-lib/bean"
	hpMessage "hp-lib/message"
	"hp-lib/net/connect"
	"hp-lib/protol"
	"log"
	"net"
	"strconv"
	"sync"
)

// 远程ID，通讯数据流
var WNConnGroup = sync.Map{}

type HpClientHandler struct {
	Key          string
	MessageType  hpMessage.HpMessage_MessageType
	ProxyAddress string
	ProxyPort    int
	CallMsg      func(message string)
	Conn         quic.Connection
	Active       bool
	InLimit      *rate.Limiter
	OutLimit     *rate.Limiter
}

// ChannelActive 连接激活时，发送注册信息给云端
func (h *HpClientHandler) ChannelActive(conn quic.Connection) {
	h.Conn = conn
	h.Active = true
	message := &hpMessage.HpMessage{
		Type: hpMessage.HpMessage_REGISTER,
		MetaData: &hpMessage.HpMessage_MetaData{
			Key: h.Key,
		},
	}
	message.MetaData.Type = h.MessageType
	stream, err := conn.OpenStream()
	if err != nil {
		h.CallMsg("获取流错误")
		return
	}
	_, err = stream.Write(protol.Encode(message))
	if err != nil {
		h.CallMsg("连接穿透服务发送数据错误：" + err.Error())
		return
	} else {
		h.CallMsg(h.ProxyAddress + ":" + strconv.Itoa(h.ProxyPort) + " 映射请求已经提交等待云端响应，请稍等")
		stream.Close()
	}
}

func (h *HpClientHandler) ChannelRead(stream quic.Stream, data interface{}) {
	message := data.(*hpMessage.HpMessage)
	switch message.Type {
	case hpMessage.HpMessage_REGISTER_RESULT:
		h.CallMsg(message.MetaData.Reason)
		break
	case hpMessage.HpMessage_CONNECTED:
		h.connected(stream, message)
		break
	case hpMessage.HpMessage_DISCONNECTED:
		if len(message.MetaData.Reason) > 0 {
			h.CallMsg(message.MetaData.Reason)
		}
		h.Close(message.MetaData.ChannelId)
		break
	case hpMessage.HpMessage_DATA:
		h.WriteData(stream, message)
	case hpMessage.HpMessage_KEEPALIVE:
		h.CallMsg("服务器端返回心跳数据")
		stream.Write(protol.Encode(&hpMessage.HpMessage{Type: hpMessage.HpMessage_KEEPALIVE}))
		break
	default:
		marshal, _ := json.Marshal(message)
		h.CallMsg("未知类型数据：" + string(marshal))
	}
}

func (h *HpClientHandler) ChannelInactive(stream quic.Stream) {
	if stream != nil {
		stream.Close()
	} else {
		h.Conn.CloseWithError(0, "关闭")
		h.Active = false
	}
}

// connected 创建内网的独立连接隧道，同时外网也重新建立一个新的
func (h *HpClientHandler) connected(stream quic.Stream, message *hpMessage.HpMessage) {
	//如果是TCP数据包，我们就连接本地的TCP服务器
	//创建外网的新连接通道
	id := message.MetaData.ChannelId
	n := &bean.WtoN{ChannelId: id, W: stream}
	WNConnGroup.Store(id, n)
	if message.MetaData.Type == hpMessage.HpMessage_TCP {
		//创建内网的新连接通道，两个实现绑定关系
		local := connect.NewTcpConnection().ConnectLocal(h.ProxyAddress, h.ProxyPort, &LocalProxyHandler{
			HpClientHandler: h,
			WToN:            n,
		}, h.CallMsg)

		if local == nil {
			h.Close(message.MetaData.ChannelId)
		}

	}
	if message.MetaData.Type == hpMessage.HpMessage_UDP {

		log.Printf("连接UDP" + message.MetaData.ChannelId)

		conn := connect.NewUdpConnection().Connect(h.ProxyAddress, h.ProxyPort, &LocalProxyUdpHandler{
			HpClientHandler: h,
			WToN:            n,
		}, h.CallMsg)
		if conn == nil {
			h.Close(message.MetaData.ChannelId)
		}
	}
}

// CloseAll 关闭所有的内网的连接通道
func (h *HpClientHandler) CloseAll() {
	WNConnGroup.Range(func(key, value interface{}) bool {
		closeHandler(key, value)
		return true
	})
}

func closeHandler(key, value interface{}) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	wToN := value.(*bean.WtoN)
	if wToN != nil {
		if wToN.N != nil {
			wToN.N.Close()
		}
		if wToN.W != nil {
			wToN.W.Close()
		}
		WNConnGroup.Delete(wToN.ChannelId)
	}
}

// Close 删除内网的连接通道
func (h *HpClientHandler) Close(channelId string) {
	load, ok := WNConnGroup.Load(channelId)
	if ok {
		wToN := load.(*bean.WtoN)
		if wToN != nil {
			if wToN.N != nil {
				wToN.N.Close()
			}
			if wToN.W != nil {
				wToN.W.Write(protol.Encode(&hpMessage.HpMessage{Type: hpMessage.HpMessage_DISCONNECTED, MetaData: &hpMessage.HpMessage_MetaData{ChannelId: channelId}}))
				wToN.W.Close()
			}
			WNConnGroup.Delete(wToN.ChannelId)
		}
	}
}

// writeData 往内网写数据
func (h *HpClientHandler) WriteData(stream quic.Stream, message *hpMessage.HpMessage) {
	log.Printf(message.MetaData.Type.String())
	load, ok := WNConnGroup.Load(message.MetaData.ChannelId)
	if !ok {
		println("不存在通道" + message.MetaData.ChannelId)
		return
	}

	wToN := load.(*bean.WtoN)
	if wToN == nil {
		return
	}
	h.writeInData(wToN.N, message.Data)
}

// writeInData 往内网写数据
func (h HpClientHandler) writeInData(conn net.Conn, data []byte) {
	if conn == nil {
		return
	}
	if h.InLimit != nil {
		b := h.InLimit.Burst()
		for {
			end := len(data)
			if end == 0 {
				break
			}
			if b < len(data) {
				end = b
			}
			err := h.InLimit.WaitN(context.Background(), end)
			if err != nil {
				h.CallMsg("往内网写数据错误：" + err.Error())
				return
			}
			_, err = conn.Write(data[:end])
			if err != nil {
				h.CallMsg("往内网写数据错误：" + err.Error())
				return
			}
			data = data[end:]
		}
	} else {
		_, err := conn.Write(data)
		if err != nil {
			h.CallMsg("往内网写数据错误：" + err.Error())
			return
		}

	}
}

// writeOutData 往外网写数据
func (h *HpClientHandler) writeOutData(stream quic.Stream, message []byte) error {
	if h.OutLimit != nil {
		b := h.OutLimit.Burst()
		for {
			end := len(message)
			if end == 0 {
				break
			}
			if b < len(message) {
				end = b
			}
			err := h.OutLimit.WaitN(context.Background(), end)
			if err != nil {
				return err
			}
			_, err = stream.Write(message[:end])
			if err != nil {
				return err
			}
			message = message[end:]
		}
	} else {
		_, err := stream.Write(message)
		return err
	}
	return nil
}
