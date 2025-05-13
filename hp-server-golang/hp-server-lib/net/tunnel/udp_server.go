package tunnel

import (
	"github.com/xtaci/smux"
	"log"
	"net"
	"strconv"
	"sync"
)

type UdpServer struct {
	cache   sync.Map
	session *smux.Session
	udpConn *net.UDPConn
}

func NewUdpServer(session *smux.Session) *UdpServer {
	return &UdpServer{
		sync.Map{},
		session,
		nil,
	}
}

// ConnectLocal 内网服务的TCP链接
func (udpServer *UdpServer) StartServer(port int) bool {
	udpAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(port))
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Printf("不能创建UDP服务器：" + ":" + strconv.Itoa(port) + " 原因：" + err.Error())
		return false
	}
	udpServer.udpConn = conn
	//设置读
	go func() {
		// 创建缓冲区用于接收数据
		buffer := make([]byte, 1450)
		for {
			if udpServer.session == nil {
				break
			}
			n, addr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				udpServer.cache.Range(func(key, value any) bool {
					handler := value.(*UdpHandler)
					go handler.ChannelInactive(conn)
					udpServer.cache.Delete(key)
					return true
				})
				break
			}
			value, ok := udpServer.cache.Load(addr.String())
			if !ok {
				handler := NewUdpHandler(udpServer, conn, udpServer.session, addr)
				go handler.ChannelActive(conn)
				udpServer.cache.Store(addr.String(), handler)
			} else {
				handler := value.(*UdpHandler)
				go handler.ChannelRead(conn, buffer[:n])
			}
		}

		udpServer.cache.Range(func(key, value any) bool {
			handler := value.(*UdpHandler)
			go handler.ChannelInactive(conn)
			udpServer.cache.Delete(key)
			return true
		})

	}()
	return true
}

func (udpServer *UdpServer) CLose() {
	if udpServer.udpConn != nil {
		udpServer.udpConn.Close()
		udpServer.udpConn = nil
	}
}
