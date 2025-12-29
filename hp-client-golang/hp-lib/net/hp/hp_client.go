package hp

import (
	"hp-lib/bean"
	hpMessage "hp-lib/message"
	"hp-lib/net"
	"hp-lib/net/connect"
	handler2 "hp-lib/net/handler"
	"hp-lib/protol"
	"sync"

	"github.com/quic-go/quic-go"
	"github.com/xtaci/smux"
	"golang.org/x/time/rate"
)

type HpClient struct {
	quit       chan struct{}
	CallMsg    func(message string)
	conn       *net.MuxSession
	quicStream *quic.Stream
	tcpStream  *smux.Stream
	syncLock   sync.Mutex
	Data       *bean.LocalInnerWear
	handler    *handler2.HpClientHandler
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
	hpClient.syncLock.Lock()
	defer hpClient.syncLock.Unlock() // 确保锁最终释放
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
	if hpClient.conn != nil {
		hpClient.conn.Close()
		hpClient.handler.CloseAll()
		hpClient.conn = nil
	}
}
