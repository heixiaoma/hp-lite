package tunnel

import (
	"bufio"
	"hp-server-lib/bean"
	net2 "hp-server-lib/net/base"
	"hp-server-lib/util"
	"log"
	"net"
	"strconv"
)

type TcpServer struct {
	conn     *net2.MuxSession
	listener net.Listener
	userInfo bean.UserConfigInfo
	limiter  *ConnectionLimiter
}

func NewTcpServer(conn *net2.MuxSession, userInfo bean.UserConfigInfo) *TcpServer {
	t := &TcpServer{
		conn,
		nil,
		userInfo,
		nil,
	}

	if userInfo.MaxConn > 0 {
		t.limiter = NewConnectionLimiter(userInfo.MaxConn)
	}
	return t
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
	// 尝试获取连接许可
	if tcpServer.limiter != nil && !tcpServer.limiter.Acquire() {
		conn.Close()
		return
	}
	go func() {

		defer func() {
			if tcpServer.limiter != nil {
				tcpServer.limiter.Release()
			}
			conn.Close()
		}()

		ip := util.GetClientIP(conn)
		if len(tcpServer.userInfo.AllowedIps) > 0 {
			ips := tcpServer.userInfo.AllowedIps
			flag := true
			for _, item := range ips {
				if util.IsIPInCIDR(ip, item) {
					flag = false
					break
				}
			}
			if flag {
				return
			}
		}

		if len(tcpServer.userInfo.BlockedIps) > 0 {
			ips := tcpServer.userInfo.BlockedIps
			for _, item := range ips {
				if util.IsIPInCIDR(ip, item) {
					return
				}
			}
		}

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
				log.Println("TCP解码错误:" + e.Error())
				handler.ChannelInactive(conn)
				return
			}
			if decode != nil {
				err := handler.ChannelRead(conn, decode)
				if err != nil {
					log.Println("TCP发送内网端错误:" + err.Error())
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
