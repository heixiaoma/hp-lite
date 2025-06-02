package hp

import (
	"golang.org/x/time/rate"
	"hp-lib/bean"
	hpMessage "hp-lib/message"
	"hp-lib/net"
	"hp-lib/net/connect"
	handler2 "hp-lib/net/handler"
	"hp-lib/protol"
	"strconv"
)

type HpClient struct {
	CallMsg func(message string)
	conn    *net.MuxSession
	Data    *bean.LocalInnerWear
	handler *handler2.HpClientHandler
}

func NewHpClient(callMsg func(message string)) *HpClient {
	return &HpClient{
		CallMsg: callMsg,
	}
}

func (hpClient *HpClient) Connect(data *bean.LocalInnerWear) {
	if hpClient.conn != nil {
		hpClient.conn.Close()
	}
	var hpType hpMessage.HpMessage_MessageType
	switch data.ConnectType {
	case bean.TCP:
		hpType = hpMessage.HpMessage_TCP
	case bean.UDP:
		hpType = hpMessage.HpMessage_UDP
	case bean.TCP_UDP:
		hpType = hpMessage.HpMessage_TCP_UDP
	}
	handler := &handler2.HpClientHandler{
		Key:          data.ConfigKey,
		MessageType:  hpType,
		ProxyAddress: data.LocalIp,
		ProxyPort:    data.LocalPort,
		CallMsg:      hpClient.CallMsg,
	}
	//限速测试
	if data.InLimit > 0 {
		handler.InLimit = rate.NewLimiter(rate.Limit(float64(data.InLimit)), data.InLimit)
	}
	if data.OutLimit > 0 {
		handler.OutLimit = rate.NewLimiter(rate.Limit(float64(data.OutLimit)), data.OutLimit)
	}
	hpClient.Data = data
	hpClient.handler = handler
	if data.TunType == "TCP" {
		connection := connect.NewHpTcpConnection()
		hpClient.conn = connection.ConnectHpTcp(data.ServerIp, data.ServerPort, handler, hpClient.CallMsg)
	} else {
		connection := connect.NewHpQuicConnection()
		hpClient.conn = connection.ConnectHpQuic(data.ServerIp, data.ServerPort, handler, hpClient.CallMsg)
	}
}

func (hpClient *HpClient) GetStatus() bool {
	if hpClient.handler != nil && hpClient.conn != nil {
		if hpClient.conn.IsTcp() {
			if hpClient.conn.TcpSession != nil {
				stream, err := hpClient.conn.TcpSession.OpenStream()
				if err != nil {
					hpClient.CallMsg("创建TCP检查流失败:" + err.Error())
					return false
				}
				_, err = stream.Write(protol.Encode(&hpMessage.HpMessage{Type: hpMessage.HpMessage_KEEPALIVE}))
				if err != nil {
					hpClient.CallMsg("TCP发送心跳包错误:" + err.Error())
					stream.Close()
					return false
				}
				err = stream.Close()
				if err != nil {
					hpClient.CallMsg("TCP关闭检查流失败:" + err.Error())
				}
				return true
			} else {
				return false
			}
		} else {
			stream, err := hpClient.conn.QuicSession.OpenUniStream()
			if err != nil {
				hpClient.CallMsg("创建QUIC检查流失败:" + err.Error())
				return false
			}
			_, err = stream.Write(protol.Encode(&hpMessage.HpMessage{Type: hpMessage.HpMessage_KEEPALIVE}))
			if err != nil {
				hpClient.CallMsg("QUIC发送心跳包错误:" + err.Error())
				stream.Close()
				return false
			}
			err = stream.Close()
			if err != nil {
				hpClient.CallMsg("QUIC关闭检查流失败:" + err.Error())
			}
			return true
		}
	} else {
		return false
	}
}

func (hpClient *HpClient) GetProxyServer() string {
	return hpClient.handler.ProxyAddress + ":" + strconv.Itoa(hpClient.handler.ProxyPort)
}

func (hpClient *HpClient) GetServer() string {
	return hpClient.Data.ServerIp + ":" + strconv.Itoa(hpClient.Data.ServerPort)
}

func (hpClient *HpClient) Close() {
	if hpClient.conn != nil {
		hpClient.conn.Close()
		hpClient.handler.CloseAll()
		hpClient.conn = nil
	}
}
