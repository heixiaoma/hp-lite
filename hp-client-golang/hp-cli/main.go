package main

import (
	"flag"
	"hp-lib/net/cmd"
	"hp-lib/util"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	var deviceId string
	var server string
	//命令行参数模式
	flag.StringVar(&deviceId, "deviceId", "", "设备ID")
	flag.StringVar(&server, "server", "", "穿透服务")
	flag.Parse()
	//默认命令行参数大于环境变量参数
	e1 := os.Getenv("deviceId")
	e2 := os.Getenv("server")
	if deviceId == "" && e1 != "" {
		deviceId = e1
	}
	if server == "" && e2 != "" {
		server = e2
	}
	if server == "" {
		log.Printf("启动失败-缺少服务启动参数 -server=xxxx.com:6666")
		for {
			time.Sleep(time.Duration(10) * time.Second)
		}
	}
	if deviceId == "" {
		log.Printf("启动失败-缺少设备ID启动参数 -deviceId=*****")
		for {
			time.Sleep(time.Duration(10) * time.Second)
		}
	}
	split := strings.Split(server, ":")
	serverPort, _ := strconv.Atoi(split[1])
	serverIp := split[0]

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
