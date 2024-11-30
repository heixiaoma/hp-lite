package tunnel

import (
	"github.com/quic-go/quic-go"
	"hp-server-lib/bean"
)

type TunnelServer struct {
	connectType bean.ConnectType
	port        int
	tcpServer   *TcpServer
	udpServer   *UdpServer
	conn        quic.Connection
}

func NewTunnelServer(connectType bean.ConnectType, port int, conn quic.Connection) *TunnelServer {
	return &TunnelServer{connectType: connectType, port: port, conn: conn}
}

func (receiver *TunnelServer) StartServer() {
	if receiver.connectType == bean.TCP || receiver.connectType == bean.TCP_UDP {
		server := NewTcpServer(receiver.conn)
		server.StartServer(receiver.port)
		receiver.tcpServer = server
	}
	if receiver.connectType == bean.UDP || receiver.connectType == bean.TCP_UDP {
		server := NewUdpServer(receiver.conn)
		server.StartServer(receiver.port)
		receiver.udpServer = server
	}
}

func (receiver *TunnelServer) CLose() {
	if receiver.tcpServer != nil {
		receiver.tcpServer.CLose()
	}
	if receiver.udpServer != nil {
		receiver.udpServer.CLose()
	}
}
