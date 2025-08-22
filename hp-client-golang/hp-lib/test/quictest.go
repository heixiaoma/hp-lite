package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/quic-go/quic-go"
	hpMessage "hp-lib/message"
	"hp-lib/protol"
)

func main() {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"HP_PRO"},
	}
	conn, err := quic.DialAddr(context.Background(), "127.0.0.1:9091", tlsConf, nil)
	if err != nil {
		fmt.Printf(err.Error())
	}

	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		fmt.Printf(err.Error())
	}

	fmt.Printf("Client: Sending '%s'\n", "message")

	message := &hpMessage.HpMessage{
		Type: hpMessage.HpMessage_REGISTER,
		MetaData: &hpMessage.HpMessage_MetaData{
			Key: "h.Key",
		},
	}

	_, err = stream.Write(protol.Encode(message))
	if err != nil {
		fmt.Printf(err.Error())
	}

	buf := make([]byte, len("message"))
	fmt.Printf("Client: Got '%s'\n", buf)
	// 主动接受服务端发起的数据流
	stream, err = conn.AcceptStream(context.Background())
	if err != nil {

		fmt.Println("无法接受数据流:", err)
		return
	}
	defer stream.Close()

	// 读取并处理数据
	data := make([]byte, 1024)
	n, err := stream.Read(data)
	if err != nil {
		fmt.Println("读取数据错误:", err)
		return
	}

	fmt.Println("接收到的数据:", string(data[:n]))
}
