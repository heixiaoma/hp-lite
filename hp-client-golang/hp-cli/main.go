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
	var c string
	//命令行参数模式
	flag.StringVar(&c, "c", "", "连接码")
	flag.Parse()
	//默认命令行参数大于环境变量参数
	e3 := os.Getenv("c")
	if c == "" && e3 != "" {
		c = e3
	}
	if c != "" {
		log.Printf("使用连接码模式连接")
		base32 := util.DecodeFromLowerCaseBase32(strings.TrimSpace(c))
		conn := strings.Split(base32, ",")
		if len(conn) != 2 {
			log.Printf("连接码错误")
			return
		}
		server := conn[0]
		deviceId := conn[1]
		split := strings.Split(server, ":")
		if len(split) != 2 {
			log.Printf("连接码错误")
			return
		}
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
}
