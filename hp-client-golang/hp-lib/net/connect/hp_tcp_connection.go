package connect

import (
	"bufio"
	"github.com/xtaci/smux"
	net2 "hp-lib/net"
	"hp-lib/protol"
	"net"
	"strconv"
)

type HpTcpConnection struct {
	Enc bool
}

func NewHpTcpConnection() *HpTcpConnection {
	return &HpTcpConnection{}
}

func (connection *HpTcpConnection) ConnectHpTcp(host string, port int, handler net2.HpHandler, call func(mgs string)) *net2.MuxSession {
	conn, err := net.Dial("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		call("不能能连到映射服务器：" + host + ":" + strconv.Itoa(port) + " 原因：" + err.Error())
		return nil
	}

	session, err := smux.Client(conn, nil)
	if err != nil {
		call("不能能连到映射服务器：" + host + ":" + strconv.Itoa(port) + " 原因：" + err.Error())
		return nil
	}
	session2 := net2.NewTcpMuxSession(session)
	handler.ChannelActive(session2)
	go func() {
		for {
			stream, err := session.AcceptStream()
			if err != nil {
				call(err.Error())
				handler.ChannelInactive(net2.NewTcpMuxStream(stream))
				return
			}
			go func() {
				reader := bufio.NewReader(stream)
				//避坑点：多包问题，需要重复读取解包
				for {
					decode, e := protol.Decode(reader)
					if e != nil {
						handler.ChannelInactive(net2.NewTcpMuxStream(stream))
						return
					}
					if decode != nil {
						handler.ChannelRead(net2.NewTcpMuxStream(stream), decode)
					}
				}
			}()
		}
	}()
	//设置读
	return session2
}
