package hp_android_lib

import (
	"hp-lib/net/cmd"
	"hp-lib/util"
	"log"
	"strconv"
	"strings"
	"time"
)

type Callback interface {
	SendResult(msg string)
}

var cmdClient *cmd.CmdClient

func Start(c string, callback Callback) {
	if c != "" {
		log.Printf("使用连接码模式连接")
		base32 := util.DecodeFromLowerCaseBase32(strings.TrimSpace(c))
		log.Printf(base32)
		con := strings.Split(base32, ",")
		server := con[0]
		deviceId := con[1]
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
	} else {
		callback.SendResult("连接码错误")
	}
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
