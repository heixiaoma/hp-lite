package server

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"github.com/quic-go/quic-go"
	net2 "hp-server-lib/net/base"
	"hp-server-lib/protol"
	"log"
	"math/big"
	"strconv"
)

type QuicServer struct {
	net2.QuicHandler
	listener *quic.Listener
}

func NewQuicServer(handler net2.QuicHandler) *QuicServer {
	return &QuicServer{
		handler,
		nil,
	}
}

func (quicServer *QuicServer) generateTLSConfig() *tls.Config {
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
func (quicServer *QuicServer) startServer(port int) {

	listener, err := quic.ListenAddr(":"+strconv.Itoa(port), quicServer.generateTLSConfig(), nil)
	if err != nil {
		log.Println("不能创建QUIC服务器：" + ":" + strconv.Itoa(port) + " 原因：" + err.Error() + " 提示：" + err.Error())
	}
	quicServer.listener = listener
	//设置读
	go func() {
		for {
			conn, err := listener.Accept(context.Background())
			if err != nil {
				log.Println("QUIC获取连接错误：" + err.Error())
			}
			go func() {
				for {
					stream, err := conn.AcceptStream(context.Background())
					if err != nil {
						quicServer.ChannelInactive(stream)
						stream.Close()
						continue
					}
					// 为每个连接启动一个新的处理 goroutine
					quicServer.handler(stream)
				}
			}()
		}
	}()
}

func (quicServer *QuicServer) handler(conn quic.Stream) {
	go func() {
		defer conn.Close()
		quicServer.ChannelActive(conn)
		reader := bufio.NewReader(conn)
		//避坑点：多包问题，需要重复读取解包
		for {
			decode, e := protol.Decode(reader)
			if e != nil {
				quicServer.ChannelInactive(conn)
				return
			}
			if decode != nil {
				quicServer.ChannelRead(conn, decode)
			}
		}
	}()
}

func (quicServer *QuicServer) CLose() {
	if quicServer.listener != nil {
		quicServer.CLose()
		quicServer.listener = nil
	}
}
