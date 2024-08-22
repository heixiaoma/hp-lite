package connect

import (
	"bufio"
	"context"
	"crypto/tls"
	"github.com/quic-go/quic-go"
	net2 "hp-lib/net"
	"hp-lib/protol"
	"os"
	"strconv"
	"strings"
	"time"
)

type QuicConnection struct {
	Enc bool
}

func NewQuicConnection() *QuicConnection {
	return &QuicConnection{}
}

func (connection *QuicConnection) ConnectHp(host string, port int, handler net2.QuicHandler, call func(mgs string)) quic.Connection {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"HP_LITE"},
	}
	q := &quic.Config{
		//最大空闲时间，超过就重连
		MaxIdleTimeout:        time.Duration(20) * time.Second,
		MaxIncomingStreams:    1000000,
		MaxIncomingUniStreams: 1000000,
		//空闲时，应该发送心跳包
		KeepAlivePeriod: time.Duration(5) * time.Second,
	}
	ctx := context.Background()
	conn, err := quic.DialAddrEarly(ctx, host+":"+strconv.Itoa(port), tlsConf, q)
	if err != nil {
		call("不能能连到映射服务器：" + host + ":" + strconv.Itoa(port) + " 原因：" + err.Error())
		//兼容模式：enc问题
		if strings.Contains(err.Error(), "invalid argument") && !connection.Enc {
			connection.Enc = true
			os.Setenv("QUIC_GO_DISABLE_ECN", "true")
			call("启用ECN禁用-兼容模式")
			return nil
		}
		if strings.Contains(err.Error(), "invalid argument") && connection.Enc {
			//最新版对GSO兼容适配做得很好了！
			os.Setenv("QUIC_GO_DISABLE_GSO", "true")
			call("启用GSO禁用-兼容模式")
		}
		return nil
	}
	handler.ChannelActive(conn)
	go func() {
		for {
			stream, err := conn.AcceptStream(context.Background())
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
	return conn
}
