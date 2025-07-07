package tunnel

import (
	"hp-server-lib/bean"
	net2 "hp-server-lib/net/base"
	"log"
)

type TunnelServer struct {
	connectType bean.ConnectType
	port        int
	tcpServer   *TcpServer
	udpServer   *UdpServer
	conn        *net2.MuxSession
	userInfo    bean.UserConfigInfo
}

func NewTunnelServer(connectType bean.ConnectType, port int, conn *net2.MuxSession, userInfo bean.UserConfigInfo) *TunnelServer {
	return &TunnelServer{connectType: connectType, port: port, conn: conn, userInfo: userInfo}
}

func (receiver *TunnelServer) StartServer() bool {

	if receiver.connectType == bean.TCP || receiver.connectType == bean.TCP_UDP {
		server := NewTcpServer(receiver.conn, receiver.userInfo)
		receiver.tcpServer = server
		if !server.StartServer(receiver.port) {
			return false
		}

	}
	if receiver.connectType == bean.UDP || receiver.connectType == bean.TCP_UDP {
		server := NewUdpServer(receiver.conn, receiver.userInfo)
		receiver.udpServer = server
		startServer := server.StartServer(receiver.port)
		if !startServer {
			return false
		}
	}
	return true
}

func (receiver *TunnelServer) CLose() {
	if receiver.tcpServer != nil {
		receiver.tcpServer.CLose()
		log.Printf("关闭TCP服务,端口：%d", receiver.port)
	}
	if receiver.udpServer != nil {
		receiver.udpServer.CLose()
		log.Printf("关闭UDP服务,端口：%d", receiver.port)
	}
}
