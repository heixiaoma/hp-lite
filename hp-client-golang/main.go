package main

import (
	"flag"
	"hp-client-golang/tcp/cmd"
	"hp-client-golang/util"
	"log"
	"os"
	"time"
)

func main() {
	var deviceId string
	//命令行参数模式
	flag.StringVar(&deviceId, "deviceId", "", "设备ID")
	flag.Parse()
	//默认命令行参数大于环境变量参数
	e := os.Getenv("deviceId")
	if deviceId == "" && e != "" {
		deviceId = e
	}
	if deviceId == "" {
		log.Printf("启动失败-缺少设备ID启动参数 -deviceId=*****")
		for {
			time.Sleep(time.Duration(10) * time.Second)
		}
	}
	serverIp := "hpproxy.cn"
	serverPort := 6666
	cmdClient := cmd.NewCmdClient(func(message string) {
		log.Printf(message)
	})
	cmdClient.Connect(serverIp, serverPort, deviceId)
	go func() {
		for {
			if !cmdClient.GetStatus() {
				cmdClient.Connect(serverIp, serverPort, deviceId)
				util.Print("中心服务器重连中")
			}
			time.Sleep(time.Duration(10) * time.Second)
		}
	}()
	select {}

}
