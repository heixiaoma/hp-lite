package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	net2 "hp-server-lib/net/base"
	"hp-server-lib/protol"
	"math/big"
	"strconv"

	"github.com/quic-go/quic-go"
)

type HpQuicServer struct {
	net2.HpHandler
	listener *quic.Listener
}

func NewHpQuicServer(handler net2.HpHandler) *HpQuicServer {
	return &HpQuicServer{
		handler,
		nil,
	}
}

func (quicServer *HpQuicServer) generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"HP_LITE"},
	}
}

// ConnectLocal 内网服务的TCP链接
func (quicServer *HpQuicServer) StartServer(port int) {
	q := &quic.Config{
		//最大空闲时间，超过就重连
		//MaxIdleTimeout:        time.Duration(20) * time.Second,
		MaxIncomingStreams:    1000000,
		MaxIncomingUniStreams: 1000000,
		//空闲时，应该发送心跳包
		//KeepAlivePeriod: time.Duration(5) * time.Second,
		Allow0RTT: true,
	}
	listener, err := quic.ListenAddr(":"+strconv.Itoa(port), quicServer.generateTLSConfig(), q)
	if err != nil {
		log.Error("不能创建UDP隧道服务器：" + ":" + strconv.Itoa(port) + " 原因：" + err.Error() + " 提示：" + err.Error())
	}
	quicServer.listener = listener
	//设置读
	go func() {
		for {
			conn, err := listener.Accept(context.Background())
			if err != nil {
				log.Error("QUIC获取连接错误：" + err.Error())
			}
			go func() {
				for {
					stream, err := conn.AcceptStream(context.Background())
					if err != nil {
						go quicServer.ChannelInactive(net2.NewQuicMuxStream(stream), net2.NewQuicMuxSession(conn))
						log.Error("接收流错误：全部关闭:", err.Error())
						return
					}
					// 为每个连接启动一个新的处理 goroutine
					quicServer.handler(net2.NewQuicMuxStream(stream), net2.NewQuicMuxSession(conn))
				}
			}()
		}
	}()
	log.Info("数据传输服务启动成功UDP:", port)
}

func (quicServer *HpQuicServer) handler(stream *net2.MuxStream, conn *net2.MuxSession) {
	go func() {
		defer stream.Close()
		quicServer.ChannelActive(stream, conn)
		reader := stream.GetReader()
		//避坑点：多包问题，需要重复读取解包
		for {
			decode, e := protol.Decode(reader)
			if e != nil {
				quicServer.ChannelInactive(stream, conn)
				return
			}
			if decode != nil {
				e := quicServer.ChannelRead(stream, decode, conn)
				if e != nil {
					return
				}
			}
		}
	}()
}

func (quicServer *HpQuicServer) CLose() {
	if quicServer.listener != nil {
		quicServer.listener.Close()
		quicServer.listener = nil
	}
}
