package tcp

import (
	"bufio"
	"io"
	"net"
	"strconv"
)

type UdpConnection struct {
}

func NewUdpConnection() *UdpConnection {
	return &UdpConnection{}
}

func (connection *UdpConnection) Connect(host string, port int, handler Handler, call func(mgs string)) net.Conn {
	conn, err := net.Dial("udp", host+":"+strconv.Itoa(port))
	if err != nil {
		call("不能能连到服务器：" + host + ":" + strconv.Itoa(port) + " 原因：" + err.Error())
		return nil
	}
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
