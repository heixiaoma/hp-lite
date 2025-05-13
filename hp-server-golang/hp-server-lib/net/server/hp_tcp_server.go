package server

import (
	"bufio"
	"github.com/xtaci/smux"
	net2 "hp-server-lib/net/base"
	"hp-server-lib/protol"
	"log"
	"net"
	"strconv"
)

type HpTcpServer struct {
	net2.HpHandler
	listener net.Listener
}

func NewHPTcpServer(handler net2.HpHandler) *HpTcpServer {
	return &HpTcpServer{
		handler,
		nil,
	}
}

// ConnectLocal 内网服务的TCP链接
func (tcpServer *HpTcpServer) StartServer(port int) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Println("不能创建TCP隧道服务器：" + ":" + strconv.Itoa(port) + " 原因：" + err.Error() + " 提示：" + err.Error())
	}
	tcpServer.listener = listener

	//设置读
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("TCP获取连接错误：" + err.Error())
			}
			session, _ := smux.Server(conn, nil)
			go func() {
				for {
					stream, err := session.AcceptStream()
					if err != nil {
						go tcpServer.ChannelInactive(&net2.MuxStream{IsTcp: true, TcpStream: stream}, &net2.MuxSession{IsTcp: true, TcpSession: session})
						log.Printf("接收流错误：全部关闭:%s", err.Error())
						return
					}
					// 为每个连接启动一个新的处理 goroutine
					tcpServer.handler(&net2.MuxStream{IsTcp: true, TcpStream: stream}, &net2.MuxSession{IsTcp: true, TcpSession: session})
				}
			}()
		}
	}()
	log.Printf("数据传输服务启动成功TCP:%d", port)
}

func (quicServer *HpTcpServer) handler(stream *net2.MuxStream, session *net2.MuxSession) {
	go func() {
		defer stream.TcpStream.Close()
		quicServer.ChannelActive(stream, session)
		reader := bufio.NewReader(stream.TcpStream)
		//避坑点：多包问题，需要重复读取解包
		for {
			decode, e := protol.Decode(reader)
			if e != nil {
				quicServer.ChannelInactive(stream, session)
				return
			}
			if decode != nil {
				e := quicServer.ChannelRead(stream, decode, session)
				if e != nil {
					return
				}
			}
		}
	}()
}

func (quicServer *HpTcpServer) CLose() {
	if quicServer.listener != nil {
		quicServer.listener.Close()
		quicServer.listener = nil
	}
}
