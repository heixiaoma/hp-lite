package tunnel

import (
	"hp-server-lib/bean"
	net2 "hp-server-lib/net/base"
	"hp-server-lib/util"
	"log"
	"net"
	"strconv"
	"sync"
)

type UdpServer struct {
	cache    sync.Map
	conn     *net2.MuxSession
	udpConn  *net.UDPConn
	userInfo bean.UserConfigInfo
}

func NewUdpServer(conn *net2.MuxSession, userInfo bean.UserConfigInfo) *UdpServer {
	return &UdpServer{
		sync.Map{},
		conn,
		nil,
		userInfo,
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
			if udpServer.conn == nil {
				break
			}
			n, addr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				break
			}

			ip := util.GetClientIPFromUDP(addr)
			if len(udpServer.userInfo.AllowedIps) > 0 {
				ips := udpServer.userInfo.AllowedIps
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

			if len(udpServer.userInfo.BlockedIps) > 0 {
				ips := udpServer.userInfo.BlockedIps
				for _, item := range ips {
					if util.IsIPInCIDR(ip, item) {
						break
					}
				}
			}

			value, ok := udpServer.cache.Load(addr.String())
			if !ok {
				err, handler := NewUdpHandler(udpServer, conn, udpServer.conn, addr, udpServer.userInfo)
				if err != nil {
					log.Println(err.Error())
					break
				}
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
