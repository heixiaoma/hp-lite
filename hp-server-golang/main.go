package main

import (
	"gopkg.in/yaml.v3"
	"hp-server-lib/config"
	"hp-server-lib/net/acme"
	"hp-server-lib/net/http"
	"hp-server-lib/net/server"
	"hp-server-lib/service"
	"hp-server-lib/task"
	"hp-server-lib/web"
	"log"
	"os"
)

func main() {

	data, err := os.ReadFile("app.yml")
	if err != nil {
		log.Printf("不存在app.yml文件,请创建一个app.yml文件，并添加相关配置")
		return
	}
	err = yaml.Unmarshal(data, &config.ConfigData)
	if err != nil {
		log.Fatalf("解析app.yml文件错误: %v", err)
	}

	if config.ConfigData.Tunnel.OpenDomain {
		go http.StartHttpServer()
		go http.StartHttpsServer()
		//缓存域名配置
		go service.InitDomainCache()
		go service.InitReverseECache()
	}
	//指令控制
	tcpServer := server.NewCmdServer()
	go tcpServer.StartServer(config.ConfigData.Cmd.Port)
	//数据传输方式1
	quicServer := server.NewHpQuicServer(server.NewHPHandler())
	go quicServer.StartServer(config.ConfigData.Tunnel.Port)
	//数据传输方式2
	hpTcpServer := server.NewHPTcpServer(server.NewHPHandler())
	go hpTcpServer.StartServer(config.ConfigData.Tunnel.Port)

	//管理后台
	go web.StartWebServer(config.ConfigData.Admin.Port)
	//acme挑战
	go func() {
		err2 := acme.StartAcmeServer(config.ConfigData.Acme.Email, config.ConfigData.Acme.HttpPort)
		if err2 != nil {
			log.Printf("证书申请服务启动失败..." + err2.Error())
		} else {
			task.StartSslTask()
		}
	}()
	select {}
}
