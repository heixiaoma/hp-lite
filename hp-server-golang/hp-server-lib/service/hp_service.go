package service

import (
	"github.com/quic-go/quic-go"
	"hp-server-lib/bean"
	"hp-server-lib/message"
	"hp-server-lib/net/tunnel"
	"hp-server-lib/protol"
	"hp-server-lib/util"
	"log"
	"strconv"
	"sync"
)

// config_key->隧道服务
var HP_CACHE_TUN = sync.Map{}
var DOMAIN_USER_INFO = sync.Map{}

type HpService struct {
}

func (receiver *HpService) loadUserConfigInfo(configKey string) *bean.UserConfigInfo {
	return &bean.UserConfigInfo{
		ProxyVersion: "",
		Domain:       "op.hp.mcle.cn",
		ProxyIp:      "192.168.100.246",
		ProxyPort:    5666,
		ConfigId:     configKey,
		Port:         8765,
		Ip:           "47.109.206.174",
	}
}

func (receiver *HpService) Register(data *message.HpMessage, conn quic.Connection) {
	configkey := data.MetaData.Key
	info := receiver.loadUserConfigInfo(configkey)
	tunnelServer, ok := HP_CACHE_TUN.Load(configkey)
	if ok {
		s := tunnelServer.(*tunnel.TunnelServer)
		s.CLose()
		HP_CACHE_TUN.Delete(configkey)
		DOMAIN_USER_INFO.Delete(info.Domain)
	}
	tunnelType := data.MetaData.Type.String()
	connectType := bean.ConnectType(tunnelType)
	newTunnelServer := tunnel.NewTunnelServer(connectType, info.Port, conn, info)
	newTunnelServer.StartServer()
	log.Printf("隧道启动成功")
	HP_CACHE_TUN.Store(configkey, newTunnelServer)
	if len(info.Domain) > 0 {
		DOMAIN_USER_INFO.Store(info.Domain, info)
	}
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
