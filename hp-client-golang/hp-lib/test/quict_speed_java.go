package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/quic-go/quic-go"
	"os"
	"time"
)

func main() {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"HP_PRO"},
	}
	conn, err := quic.DialAddr(context.Background(), "127.0.0.1:9999", tlsConf, nil)
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
