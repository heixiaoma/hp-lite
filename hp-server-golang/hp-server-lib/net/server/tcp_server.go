package server

import (
	"bufio"
	net2 "hp-server-lib/net/base"
	"log"
	"net"
	"strconv"
)

type TcpServer struct {
	net2.Handler
	listener net.Listener
}

func NewTcpServer(handler net2.Handler) *TcpServer {
	return &TcpServer{
		handler,
		nil,
	}
}

// ConnectLocal 内网服务的TCP链接
func (tcpServer *TcpServer) startServer(port int) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatalf("不能创建TCP服务器：" + ":" + strconv.Itoa(port) + " 原因：" + err.Error() + " 提示：" + err.Error())
		return
	}
	tcpServer.listener = listener
	//设置读
	go func() {
		for {
			if tcpServer.listener == nil {
				return
			}
			conn, err := listener.Accept()
			if err == nil {
				tcpServer.handler(conn)
			} else {
				log.Println("TCP错误连接:", err)
			}
		}
	}()
}

func (tcpServer *TcpServer) handler(conn net.Conn) {
	go func() {
		defer conn.Close()
		tcpServer.ChannelActive(conn)
		reader := bufio.NewReader(conn)
		for {
			if tcpServer.listener == nil {
				return
			}
			//尝试读检查连接激活
			_, err := reader.Peek(1)
			if err != nil {
				tcpServer.ChannelInactive(conn)
				return
			}

			decode, e := tcpServer.Decode(reader)
			if e != nil {
				tcpServer.ChannelInactive(conn)
				return
			}
			if decode != nil {
				tcpServer.ChannelRead(conn, decode)
			}
		}
	}()
}

func (tcpServer *TcpServer) CLose() {
	if tcpServer.listener != nil {
		tcpServer.CLose()
		tcpServer.listener = nil
	}
}
