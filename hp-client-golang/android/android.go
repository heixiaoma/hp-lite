package android

import (
	"hp-client-golang/tcp/cmd"
	"strconv"
	"strings"
	"time"
)

type Callback interface {
	SendResult(msg string)
}

var cmdClient *cmd.CmdClient

// Start 创建开始时间
func Start(server string, deviceId string, callback Callback) {
	split := strings.Split(server, ":")
	serverPort, _ := strconv.Atoi(split[1])
	cmdClient = cmd.NewCmdClient(callback.SendResult)
	cmdClient.Connect(split[0], serverPort, deviceId)
	go func() {
		for {
			if !cmdClient.GetStatus() {
				cmdClient.Connect(split[0], serverPort, deviceId)
				callback.SendResult("中心服务器重连中")
			}
			time.Sleep(time.Duration(10) * time.Second)
		}
	}()
	select {}
}

// Close 关闭服务
func Close() bool {
	if cmdClient != nil {
		cmdClient.Close()
		return true
	}
	return false
}
