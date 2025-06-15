package cmd

import (
	"bufio"
	net2 "hp-lib/net"
	"hp-lib/protol"
	"net"
	"strconv"
	"sync/atomic"
)

type TcpConnection struct {
	IsClose atomic.Value
}

func NewTcpConnection() *TcpConnection {
	t := &TcpConnection{}
	t.IsClose.Store(false)
	return t
}

func (connection *TcpConnection) Connect(host string, port int, handler net2.Handler, call func(mgs string)) net.Conn {
	conn, err := net.Dial("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		call("不能能连到服务器：" + host + ":" + strconv.Itoa(port) + " 原因：" + err.Error())
		return nil
	}
	handler.ChannelActive(conn)
	//设置读
	go func() {
		reader := bufio.NewReader(conn)
		for {
			if connection.IsClosed() {
				handler.ChannelInactive(conn)
				return
			}
			//尝试读检查连接激活
			_, err := reader.Peek(1)
			if err != nil {
				handler.ChannelInactive(conn)
				return
			}
			decode, e := protol.CmdDecode(reader)
			if e != nil {
				call(e.Error())
				handler.ChannelInactive(conn)
				return
			}
			if decode != nil {
				handler.ChannelRead(conn, decode)
			}
		}
	}()
	return conn
}
func (receiver *TcpConnection) IsClosed() bool {
	return receiver.IsClose.Load().(bool)
}

func (receiver *TcpConnection) Close() {
	receiver.IsClose.Store(true)
}
