package connect

import (
	"bufio"
	net2 "hp-lib/net"
	"hp-lib/util"
	"io"
	"net"
	"strconv"
)

type UdpConnection struct {
}

func NewUdpConnection() *UdpConnection {
	return &UdpConnection{}
}

func (connection *UdpConnection) Connect(address string, handler net2.Handler, call func(mgs string)) net.Conn {
	err2, _, host, port := util.ProtocolInfo(address)
	if err2 != nil {
		call("地址解析错误：" + host + ":" + strconv.Itoa(port) + " 原因：" + err2.Error())
		return nil
	}

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
