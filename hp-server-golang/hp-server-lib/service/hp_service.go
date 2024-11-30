package service

import (
	"github.com/quic-go/quic-go"
	"hp-server-lib/bean"
	"hp-server-lib/message"
	"hp-server-lib/net/server"
	"hp-server-lib/protol"
	"hp-server-lib/util"
	"log"
	"strconv"
	"sync"
)

// config_key->隧道服务
var HP_CACHE_TUN = sync.Map{}

type HpService struct {
}

func (receiver *HpService) loadUserConfigInfo(configKey string) *bean.UserConfigInfo {
	return &bean.UserConfigInfo{
		ProxyVersion: "",
		ProxyIp:      "192.168.100.1",
		ProxyPort:    80,
		ConfigId:     configKey,
		Port:         8765,
		Ip:           "127.0.0.1",
	}

}

func (receiver *HpService) Register(stream quic.Stream, data *message.HpMessage, conn quic.Connection) {
	configkey := data.MetaData.Key
	tunnelServer, ok := CMD_CACHE_CONN.Load(configkey)
	if ok {
		s := tunnelServer.(server.TunnelServer)
		s.CLose()
		CMD_CACHE_CONN.Delete(configkey)
	}
	info := receiver.loadUserConfigInfo(configkey)
	//tunnelType := data.MetaData.Type.String()
	//connectType := bean.ConnectType(tunnelType)
	//newTunnelServer := server.NewTunnelServer(connectType, info.Port)
	//newTunnelServer.StartServer()
	log.Printf("隧道启动成功")
	CMD_CACHE_CONN.Store(data.MetaData.Key, tunnelServer)
	//通知客户端结果
	arr2 := [][]string{
		{"穿透结果", "穿透成功"},
		{"内外TCP", info.ProxyIp + ":" + strconv.Itoa(info.ProxyPort)},
		{"外网TCP", info.Ip + ":" + strconv.Itoa(info.Port)},
	}
	status := util.PrintStatus(arr2)
	m := &message.HpMessage_MetaData{
		Success: true,
		Reason:  status,
	}
	hpMessage := &message.HpMessage{
		Type:     message.HpMessage_REGISTER_RESULT,
		MetaData: m,
	}
	openStream, err := conn.OpenStream()
	if err == nil {
		openStream.Write(protol.Encode(hpMessage))
		util.Print(status)
	}
}
