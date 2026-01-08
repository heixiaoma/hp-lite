package main

import (
	"hp-server-lib/log"
	"net"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", ":18850")
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return
	}
	connc, err := net.Dial("udp", "127.0.0.1:15500")
	if err != nil {
		return
	}
	//设置读
	go func() {
		buffer := make([]byte, 8192)
		// 创建缓冲区用于接收数据
		for {
			n, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				log.Error("udp读取数据错误 原因：" + err.Error())
				break
			}
			bytes := buffer[:n]
			connc.Write(bytes)
		}

	}()
}
