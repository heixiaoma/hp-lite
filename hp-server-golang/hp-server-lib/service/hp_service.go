package service

import (
	"github.com/quic-go/quic-go"
	"hp-server-lib/bean"
	"hp-server-lib/db"
	"hp-server-lib/entity"
	"hp-server-lib/message"
	"hp-server-lib/net/tunnel"
	"hp-server-lib/protol"
	"hp-server-lib/util"
	"log"
	"strconv"
	"sync"
)

// 端口->隧道服务
var HP_CACHE_TUN = sync.Map{}
var DOMAIN_USER_INFO = sync.Map{}

type HpService struct {
}

func (receiver *HpService) loadUserConfigInfo(configKey string) *bean.UserConfigInfo {

	userQuery := &entity.UserConfigEntity{}
	db.DB.Where("config_key = ? ", configKey).First(userQuery)
	if userQuery == nil || userQuery.LocalPort == nil || userQuery.Port == nil || userQuery.Id == nil {
		return nil
	}
	s := ""
	if userQuery.ProxyVersion == bean.V1 {
		s = "V1"
	} else if userQuery.ProxyVersion == bean.V2 {
		s = "V2"
	}
	return &bean.UserConfigInfo{
		ProxyVersion:       s,
		Domain:             userQuery.Domain,
		ProxyIp:            userQuery.LocalIp,
		ProxyPort:          *userQuery.LocalPort,
		ConfigId:           *userQuery.Id,
		Port:               *userQuery.Port,
		Ip:                 userQuery.ServerIp,
		CertificateKey:     userQuery.CertificateKey,
		CertificateContent: userQuery.CertificateContent,
	}
}

func (receiver *HpService) Register(data *message.HpMessage, conn quic.Connection) {
	configkey := data.MetaData.Key
	info := receiver.loadUserConfigInfo(configkey)
	if info == nil {
		return
	}
	tunnelServer, ok := HP_CACHE_TUN.Load(info.Port)
	if ok {
		s := tunnelServer.(*tunnel.TunnelServer)
		s.CLose()
		HP_CACHE_TUN.Delete(info.Port)
		if info.Domain != nil {
			DOMAIN_USER_INFO.Delete(*info.Domain)
		}
	}
	tunnelType := data.MetaData.Type.String()
	connectType := bean.ConnectType(tunnelType)
	newTunnelServer := tunnel.NewTunnelServer(connectType, info.Port, conn, *info)
	server := newTunnelServer.StartServer()
	if !server {
		newTunnelServer.CLose()
	} else {
		log.Printf("隧道启动成功")
		HP_CACHE_TUN.Store(info.Port, newTunnelServer)
	}
	if info.Domain != nil {
		DOMAIN_USER_INFO.Store(*info.Domain, info)
	}

	//更新服务端状态
	strMsg := "配置启动成功"
	if !server {
		strMsg = "配置启动失败，大概率是端口冲突，请刷新"
	}
	db.DB.Model(&entity.UserConfigEntity{}).Where("config_key = ?", configkey).Update("status_msg", strMsg)
	//通知客户端结果
	arr2 := [][]string{
		{"穿透结果", strconv.FormatBool(server)},
	}

	if server && (connectType == bean.TCP || connectType == bean.TCP_UDP) {
		arr2 = append(arr2, []string{"内网TCP", info.ProxyIp + ":" + strconv.Itoa(info.ProxyPort)})
		arr2 = append(arr2, []string{"外网TCP", info.Ip + ":" + strconv.Itoa(info.Port)})
		if info.Domain != nil {
			arr2 = append(arr2, []string{"HTTP地址", "http://" + *info.Domain})
		}
		if len(info.CertificateKey) > 0 && len(info.CertificateContent) > 0 {
			arr2 = append(arr2, []string{"HTTPS地址", "https://" + *info.Domain})
		}
	}

	if server && (connectType == bean.UDP || connectType == bean.TCP_UDP) {
		arr2 = append(arr2, []string{"内网UDP", info.ProxyIp + ":" + strconv.Itoa(info.ProxyPort)})
		arr2 = append(arr2, []string{"外网UDP", info.Ip + ":" + strconv.Itoa(info.Port)})
	}

	status := util.PrintStatus(arr2)
	m := &message.HpMessage_MetaData{
		Success: server,
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
