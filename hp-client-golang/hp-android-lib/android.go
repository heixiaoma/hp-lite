package hp_android_lib

import (
	"hp-lib/net/cmd"
	"strconv"
	"strings"
	"time"
)

type Callback interface {
	SendResult(msg string)
}

var cmdClient *cmd.CmdClient

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

func Close() bool {
	if cmdClient != nil {
		cmdClient.Close()
		return true
	}
	return false
}

func GetStatus() bool {
	if cmdClient != nil {
		return cmdClient.GetStatus()
	}
	return false
}
