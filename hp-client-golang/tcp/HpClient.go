package tcp

import (
	"hp-client-golang/bean"
	"hp-client-golang/hpMessage"
	"net"
	"strconv"
)

type HpClient struct {
	CallMsg func(message string)
	conn    net.Conn
	Data    *bean.LocalInnerWear
	handler *HpClientHandler
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
	connection := NewTcpConnection()
	var hpType hpMessage.HpMessage_MessageType
	switch data.ConnectType {
	case bean.TCP:
		hpType = hpMessage.HpMessage_TCP
	case bean.UDP:
		hpType = hpMessage.HpMessage_UDP
	case bean.TCP_UDP:
		hpType = hpMessage.HpMessage_TCP_UDP
	}

	handler := &HpClientHandler{
		Key:          data.ConfigKey,
		MessageType:  hpType,
		ProxyAddress: data.LocalIp,
		ProxyPort:    data.LocalPort,
		CallMsg:      hpClient.CallMsg,
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
		hpClient.conn.Close()
	}
	if hpClient.handler != nil {
		hpClient.handler.CloseAll()
	}
}
