package hp

import (
	"github.com/quic-go/quic-go"
	"golang.org/x/time/rate"
	"hp-lib/bean"
	hpMessage "hp-lib/message"
	"hp-lib/net/connect"
	handler2 "hp-lib/net/handler"
	"strconv"
)

type HpClient struct {
	CallMsg func(message string)
	conn    quic.Connection
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
		hpClient.conn.CloseWithError(0, "重连关闭")
	}
	connection := connect.NewQuicConnection()
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
	hpClient.conn = connection.ConnectHp(data.ServerIp, data.ServerPort, handler, hpClient.CallMsg)
}

func (hpClient *HpClient) GetStatus() bool {
	if hpClient.handler != nil {
		return hpClient.handler.Active
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
		hpClient.conn.CloseWithError(0, "正常关闭")
		hpClient.handler.CloseAll()
		hpClient.conn = nil
	}
}
