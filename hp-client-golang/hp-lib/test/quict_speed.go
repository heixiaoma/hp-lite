package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/quic-go/quic-go"
	"io"
	"math/big"
	"os"
	"time"
)

const addr = "localhost:4242"

const message = "foobar"

func generateTLSConfig() *tls.Config {
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
		NextProtos:   []string{"HP_PRO"},
	}
}
func server() {
	listener, err := quic.ListenAddr(addr, generateTLSConfig(), nil)
	if err != nil {
		println(err.Error())
	}
	conn, err := listener.Accept(context.Background())
	if err != nil {
		println(err.Error())

	}
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		println(err.Error())
	}
	reader := bufio.NewReader(stream)
	start := time.Now()

	for {
		//尝试读检查连接激活
		_, err := reader.Peek(1)
		if err != nil {
			break
		}
		data := make([]byte, reader.Buffered())
		io.ReadFull(reader, data)
	}
	elapsed := time.Since(start)
	fmt.Printf("接收耗时：%s\n", elapsed)
}

func client() {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"HP_PRO"},
	}
	conn, err := quic.DialAddr(context.Background(), addr, tlsConf, nil)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		fmt.Printf(err.Error())
	}

	fmt.Printf("Client: Sending '%s'\n", "message")

	file, err := os.Open("/Users/heixiaoma/Downloads/阴阳画皮.mkv.7")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	buffer := make([]byte, 2048)
	start := time.Now()
	for {
		bytesRead, err := file.Read(buffer)
		if err != nil {
			fmt.Println(err)
			break
		}
		if bytesRead == 0 {
			break
		}
		_, err = stream.Write(buffer)
		if err != nil {
			fmt.Println(err)
			break
		}
	}
	elapsed := time.Since(start)
	fmt.Printf("发送耗时：%s\n", elapsed)
	stream.Close()
	if err != nil {
		fmt.Printf(err.Error())
	}

}

func main() {
	go server()
	go client()
	select {}
}
