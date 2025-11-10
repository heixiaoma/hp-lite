package connect

import (
	"bufio"
	net2 "hp-lib/net"
	"hp-lib/net/socks5"
	"hp-lib/util"
	"io"
	"net"
	"time"
)

type Socks5Connection struct {
}

func NewSocks5Connection() *Socks5Connection {
	return &Socks5Connection{}
}

func (connection *Socks5Connection) ConnectSocks(address string, handler net2.Handler, call func(mgs string)) net.Conn {
	server := socks5.Server{}
	username, password, _ := util.ParseSocks5Auth(address)
	server.ListenBindReuseTimeout = time.Second / 2
	if username != "" && password != "" {
		server.Authentication = socks5.UserAuth(username, password)
	}
	conn, conn2 := net.Pipe()
	go server.ServeConn(conn2)
	handler.ChannelActive(conn)
	//设置读
	go func() {
		reader := bufio.NewReader(conn)
		for {
			//尝试读检查连接激活
			_, err := reader.Peek(1)
			if err != nil {
				handler.ChannelInactive(conn)
				return
			}
			if reader.Buffered() > 0 {
				data := make([]byte, reader.Buffered())
				io.ReadFull(reader, data)
				handler.ChannelRead(conn, data)
			}
		}
	}()
	return conn
}
