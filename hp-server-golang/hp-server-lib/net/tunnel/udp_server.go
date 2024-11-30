package tunnel

import (
	net2 "hp-server-lib/net/base"
	"log"
	"net"
	"strconv"
)

type UdpServer struct {
	net2.UdpHandler
	conn net.Conn
}

func NewUdpServer(handler net2.UdpHandler) *UdpServer {
	return &UdpServer{
		handler,
		nil,
	}
}

// ConnectLocal 内网服务的TCP链接
func (udpServer *UdpServer) StartServer(port int) {
	udpAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(port))
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalf("不能创建UDP服务器：" + ":" + strconv.Itoa(port) + " 原因：" + err.Error() + " 提示：" + err.Error())
		return
	}
	udpServer.conn = conn
	//设置读
	go func() {
		// 创建缓冲区用于接收数据
		buffer := make([]byte, 1024)
		for {
			if udpServer.conn == nil {
				return
			}
			n, addr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				log.Println("错误读取UDP信息:", err)
				continue
			}
			go udpServer.ChannelRead(conn, addr, buffer[:n])
		}
	}()
}

func (udpServer *UdpServer) CLose() {
	if udpServer.conn != nil {
		udpServer.conn.Close()
		udpServer.conn = nil
	}
}
