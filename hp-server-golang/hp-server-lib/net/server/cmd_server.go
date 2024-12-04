package server

import (
	"bufio"
	net2 "hp-server-lib/net/base"
	"log"
	"net"
	"strconv"
)

type CmdServer struct {
	net2.Handler
	listener net.Listener
}

func NewCmdServer(handler net2.Handler) *CmdServer {
	return &CmdServer{
		handler,
		nil,
	}
}

// ConnectLocal 内网服务的TCP链接
func (tcpServer *CmdServer) StartServer(port int) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Printf("不能创建TCP服务器：" + ":" + strconv.Itoa(port) + " 原因：" + err.Error() + " 提示：" + err.Error())
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
				log.Println("TCP错误连接222:", err)
			}
		}
	}()
	log.Printf("指令传输服务启动成功TCP:%d", port)

}

func (tcpServer *CmdServer) handler(conn net.Conn) {
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
				log.Println(e)
				tcpServer.ChannelInactive(conn)
				return
			}
			if decode != nil && conn != nil {
				tcpServer.ChannelRead(conn, decode)
			} else {
				return
			}
		}
	}()
}

func (tcpServer *CmdServer) CLose() {
	if tcpServer.listener != nil {
		tcpServer.listener.Close()
		tcpServer.listener = nil
	}
}
