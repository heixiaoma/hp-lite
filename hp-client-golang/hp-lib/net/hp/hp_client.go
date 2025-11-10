package hp

import (
	"hp-lib/bean"
	hpMessage "hp-lib/message"
	"hp-lib/net"
	"hp-lib/net/connect"
	handler2 "hp-lib/net/handler"
	net2 "hp-lib/net/socks5"
	"hp-lib/protol"
	"hp-lib/util"
	"strconv"

	"github.com/quic-go/quic-go"
	"github.com/xtaci/smux"
	"golang.org/x/time/rate"
)

type HpClient struct {
	CallMsg    func(message string)
	conn       *net.MuxSession
	quicStream *quic.Stream
	tcpStream  *smux.Stream

	Data    *bean.LocalInnerWear
	handler *handler2.HpClientHandler

	socksServer *net2.SocksServer
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
	hpClient.Data = data
	err, s, _, port := util.ProtocolInfo(data.LocalAddress)
	if err != nil {
		hpClient.CallMsg(err.Error())
		return
	}

	if s == "socks5" {
		username, password, err := util.ParseSocks5Auth(data.LocalAddress)
		if err != nil {
			hpClient.CallMsg(err.Error())
		}
		hpClient.socksServer = net2.NewSocks(strconv.Itoa(port), username, password, hpClient.CallMsg)
		hpClient.socksServer.Start(func() {
			hpClient.CallMsg("SOCKS5服务停止")
		})
	}

	handler := &handler2.HpClientHandler{
		Key:          data.ConfigKey,
		LocalAddress: data.LocalAddress,
		CallMsg:      hpClient.CallMsg,
	}
	//限速测试
	if data.InLimit > 0 {
		handler.InLimit = rate.NewLimiter(rate.Limit(float64(data.InLimit)), data.InLimit)
	}
	if data.OutLimit > 0 {
		handler.OutLimit = rate.NewLimiter(rate.Limit(float64(data.OutLimit)), data.OutLimit)
	}
	hpClient.handler = handler
	if data.TunType == "TCP" {
		connection := connect.NewHpTcpConnection()
		hpClient.tcpStream = nil
		hpClient.conn = connection.ConnectHpTcp(data.ServerIp, data.ServerPort, handler, hpClient.CallMsg)
	} else {
		connection := connect.NewHpQuicConnection()
		hpClient.quicStream = nil
		hpClient.conn = connection.ConnectHpQuic(data.ServerIp, data.ServerPort, handler, hpClient.CallMsg)
	}
}

func (hpClient *HpClient) GetStatus() bool {
	if hpClient.handler != nil && hpClient.conn != nil {
		if hpClient.conn.IsTcp() {
			if hpClient.conn.TcpSession != nil {
				if hpClient.tcpStream == nil {
					stream, err := hpClient.conn.TcpSession.OpenStream()
					if err != nil {
						hpClient.CallMsg("创建TCP检查流失败:" + err.Error())
						return false
					}
					hpClient.tcpStream = stream
				}
				_, err := hpClient.tcpStream.Write(protol.Encode(&hpMessage.HpMessage{Type: hpMessage.HpMessage_KEEPALIVE}))
				if err != nil {
					hpClient.CallMsg("TCP发送心跳包错误:" + err.Error())
					hpClient.tcpStream.Close()
					hpClient.tcpStream = nil
					return false
				}
				return true
			} else {
				return false
			}
		} else {
			if hpClient.quicStream == nil {
				stream, err := hpClient.conn.QuicSession.OpenStream()
				if err != nil {
					hpClient.CallMsg("创建QUIC检查流失败:" + err.Error())
					return false
				}
				hpClient.quicStream = stream
			}
			_, err := hpClient.quicStream.Write(protol.Encode(&hpMessage.HpMessage{Type: hpMessage.HpMessage_KEEPALIVE}))
			if err != nil {
				hpClient.CallMsg("QUIC发送心跳包错误:" + err.Error())
				hpClient.quicStream.Close()
				hpClient.quicStream = nil
				return false
			}
			return true
		}
	} else {
		return false
	}
}

func (hpClient *HpClient) Close() {
	if hpClient.socksServer != nil {
		hpClient.socksServer.Stop()
		hpClient.socksServer = nil
	}
	if hpClient.conn != nil {
		hpClient.conn.Close()
		hpClient.handler.CloseAll()
		hpClient.conn = nil
	}
}
