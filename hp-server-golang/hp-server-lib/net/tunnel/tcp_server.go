package tunnel

import (
	"bufio"
	"github.com/quic-go/quic-go"
	"hp-server-lib/bean"
	"log"
	"net"
	"strconv"
)

type TcpServer struct {
	conn     quic.Connection
	listener net.Listener
	userInfo bean.UserConfigInfo
}

func NewTcpServer(conn quic.Connection, userInfo bean.UserConfigInfo) *TcpServer {
	return &TcpServer{
		conn,
		nil,
		userInfo,
	}
}

// ConnectLocal 内网服务的TCP链接
func (tcpServer *TcpServer) StartServer(port int) bool {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Printf("不能创建TCP服务器：" + ":" + strconv.Itoa(port) + " 原因：" + err.Error())
		return false
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
			}
		}
	}()
	return true
}

func (tcpServer *TcpServer) handler(conn net.Conn) {
	go func() {
		defer conn.Close()
		handler := NewTcpHandler(conn, tcpServer.conn, tcpServer.userInfo)
		handler.ChannelActive(conn)
		reader := bufio.NewReader(conn)
		for {
			if tcpServer.listener == nil {
				return
			}
			//尝试读检查连接激活
			_, err := reader.Peek(1)
			if err != nil {
				handler.ChannelInactive(conn)
				return
			}

			decode, e := handler.Decode(reader)
			if e != nil {
				log.Println(e)
				handler.ChannelInactive(conn)
				return
			}
			if decode != nil {
				err := handler.ChannelRead(conn, decode)
				if err != nil {
					log.Println(e)
					return
				}
			}
		}
	}()
}

func (tcpServer *TcpServer) CLose() {
	if tcpServer.listener != nil {
		tcpServer.listener.Close()
		tcpServer.listener = nil
	}
}
