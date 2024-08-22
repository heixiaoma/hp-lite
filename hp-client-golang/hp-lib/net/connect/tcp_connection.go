package connect

import (
	"bufio"
	net2 "hp-lib/net"
	"io"
	"net"
	"strconv"
	"time"
)

type TcpConnection struct {
}

func NewTcpConnection() *TcpConnection {
	return &TcpConnection{}
}

// ConnectLocal 内网服务的TCP链接
func (connection *TcpConnection) ConnectLocal(host string, port int, handler net2.Handler, call func(mgs string)) net.Conn {
	// 将超时时间转换为 time.Duration 类型
	timeoutDuration := time.Duration(5) * time.Second
	conn, err := net.DialTimeout("tcp", host+":"+strconv.Itoa(port), timeoutDuration)
	if err != nil {
		call("不能能连到内网服务器：" + host + ":" + strconv.Itoa(port) + " 原因：" + err.Error() + " 提示：请检查本地服务是否正常打开")
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
