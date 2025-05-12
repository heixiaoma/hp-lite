package connect

import (
	"bufio"
	"github.com/hashicorp/yamux"
	net2 "hp-lib/net"
	"hp-lib/protol"
	"net"
	"strconv"
)

type HpConnection struct {
}

func NewQuicConnection() *HpConnection {
	return &HpConnection{}
}

func (connection *HpConnection) ConnectHp(host string, port int, handler net2.HpHandler, call func(mgs string)) *yamux.Session {
	conn, err := net.Dial("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		call("不能能连到映射服务器：" + host + ":" + strconv.Itoa(port) + " 原因：" + err.Error())
		return nil
	}

	session, err := yamux.Client(conn, nil)
	if err != nil {
		call("不能能连到映射服务器：" + host + ":" + strconv.Itoa(port) + " 原因：" + err.Error())
		return nil
	}
	handler.ChannelActive(session)
	go func() {
		for {
			stream, err := session.AcceptStream()
			if err != nil {
				call(err.Error())
				handler.ChannelInactive(stream)
				return
			}
			go func() {
				reader := bufio.NewReader(stream)
				//避坑点：多包问题，需要重复读取解包
				for {
					decode, e := protol.Decode(reader)
					if e != nil {
						handler.ChannelInactive(stream)
						return
					}
					if decode != nil {
						handler.ChannelRead(stream, decode)
					}
				}
			}()
		}
	}()
	//设置读
	return session
}
