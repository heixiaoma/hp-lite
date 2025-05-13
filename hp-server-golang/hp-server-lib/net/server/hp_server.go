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

type HpServer struct {
	net2.HpHandler
	listener net.Listener
}

func NewHPServer(handler net2.HpHandler) *HpServer {
	return &HpServer{
		handler,
		nil,
	}
}

// ConnectLocal 内网服务的TCP链接
func (hpServer *HpServer) StartServer(port int) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Println("不能创建QUIC服务器：" + ":" + strconv.Itoa(port) + " 原因：" + err.Error() + " 提示：" + err.Error())
	}
	hpServer.listener = listener

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
						go hpServer.ChannelInactive(stream, session)
						log.Printf("接收流错误：全部关闭:%s", err.Error())
						return
					}
					// 为每个连接启动一个新的处理 goroutine
					hpServer.handler(stream, session)
				}
			}()
		}
	}()
	log.Printf("数据传输服务启动成功UDP:%d", port)
}

func (quicServer *HpServer) handler(stream *smux.Stream, session *smux.Session) {
	go func() {
		defer stream.Close()
		quicServer.ChannelActive(stream, session)
		reader := bufio.NewReader(stream)
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

func (quicServer *HpServer) CLose() {
	if quicServer.listener != nil {
		quicServer.listener.Close()
		quicServer.listener = nil
	}
}
