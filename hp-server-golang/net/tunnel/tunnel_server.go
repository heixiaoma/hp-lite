package tunnel

import (
	"hp-server-lib/bean"
	net2 "hp-server-lib/net/base"
	"log"
)

type TunnelServer struct {
	protocol  bean.Protocol
	port      int
	tcpServer *TcpServer
	udpServer *UdpServer
	conn      *net2.MuxSession
	userInfo  bean.UserConfigInfo
}

func NewTunnelServer(protocol bean.Protocol, port int, conn *net2.MuxSession, userInfo bean.UserConfigInfo) *TunnelServer {
	return &TunnelServer{protocol: protocol, port: port, conn: conn, userInfo: userInfo}
}

func (receiver *TunnelServer) UserInfo() bean.UserConfigInfo {
	return receiver.userInfo
}

func (receiver *TunnelServer) StartServer() bool {

	if receiver.protocol == bean.TCP || receiver.protocol == bean.HTTP || receiver.protocol == bean.SOCKS5 || receiver.protocol == bean.HTTPS || receiver.protocol == bean.UNIX {
		server := NewTcpServer(receiver.conn, receiver.userInfo)
		receiver.tcpServer = server
		if !server.StartServer(receiver.port) {
			return false
		}

	}
	if receiver.protocol == bean.UDP {
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
