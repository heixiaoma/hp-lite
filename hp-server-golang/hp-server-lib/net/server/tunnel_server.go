package server

import (
	"hp-server-lib/bean"
	"hp-server-lib/net/handler"
)

type TunnelServer struct {
	connectType bean.ConnectType
	port        int
	tcpServer   *TcpServer
	udpServer   *UdpServer
}

func NewTunnelServer(connectType bean.ConnectType, port int) *TunnelServer {
	return &TunnelServer{connectType: connectType, port: port}
}

func (receiver *TunnelServer) StartServer() {
	if receiver.connectType == bean.TCP || receiver.connectType == bean.TCP_UDP {
		server := NewTcpServer(handler.NewTcpClientHandler())
		server.StartServer(receiver.port)
		receiver.tcpServer = server
	}
	if receiver.connectType == bean.UDP || receiver.connectType == bean.TCP_UDP {
		clientHandler := &handler.UdpClientHandler{}
		server := NewUdpServer(clientHandler)
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
