package connect

import (
	"bufio"
	"hp-lib/bean"
	net2 "hp-lib/net"
	"hp-lib/util"
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
func (connection *TcpConnection) ConnectLocal(address string, handler net2.Handler, call func(mgs string)) net.Conn {
	// 将超时时间转换为 time.Duration 类型
	timeoutDuration := time.Duration(5) * time.Second
	err2, s, host, port := util.ProtocolInfo(address)
	if err2 != nil {
		call("解析内网服务器错误：" + err2.Error() + " 提示：请检查穿透配置是否正常")
		return nil
	}
	protocol := "tcp"
	address2 := host + ":" + strconv.Itoa(port)
	if bean.Protocol(s) == bean.UNIX {
		protocol = "unix"
		address2 = host
	}
	conn, err := net.DialTimeout(protocol, address2, timeoutDuration)
	if err != nil {
		call("不能能连到内网服务器：" + address2 + " 原因：" + err.Error() + " 提示：请检查本地服务是否正常打开")
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
