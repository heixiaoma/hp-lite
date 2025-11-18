package service

import (
	"hp-server-lib/bean"
	"hp-server-lib/db"
	"hp-server-lib/entity"
	"hp-server-lib/log"
	"hp-server-lib/message"
	net2 "hp-server-lib/net/base"
	"hp-server-lib/net/tunnel"
	"hp-server-lib/protol"
	"hp-server-lib/util"
	"strconv"
	"sync"
	"time"
)

// 端口->隧道服务
var HP_CACHE_TUN = sync.Map{}
var DOMAIN_HP_INFO = sync.Map{}
var mu sync.Mutex // 创建一个互斥锁

type HpService struct {
}

func (receiver *HpService) loadUserConfigInfo(configKey string) *bean.UserConfigInfo {

	userQuery := &entity.UserConfigEntity{}
	db.DB.Where("config_key = ?  and  (status = 0 or status is null) ", configKey).First(userQuery)
	if userQuery == nil || userQuery.Id == nil {
		return nil
	}

	s := ""
	if userQuery.ProxyVersion == bean.V1 {
		s = "V1"
	} else if userQuery.ProxyVersion == bean.V2 {
		s = "V2"
	}

	b := &bean.UserConfigInfo{
		ProxyVersion: s,
		Domain:       userQuery.Domain,
		LocalAddress: userQuery.LocalAddress,
		ConfigId:     *userQuery.Id,
		RemotePort:   *userQuery.RemotePort,
		Ip:           userQuery.ServerIp,
		TunType:      userQuery.TunType,
		MaxConn:      -1,
	}
	//防火墙参数配置
	waf := &entity.UserWafEntity{}
	db.DB.Where("config_id = ?", userQuery.Id).First(waf)
	if waf != nil {
		if len(waf.AllowedIPs) > 0 {
			b.AllowedIps = waf.AllowedIPs
		}
		if len(waf.BlockedIPs) > 0 {
			b.BlockedIps = waf.BlockedIPs
		}
		if waf.RateLimit > 0 {
			b.MaxConn = waf.RateLimit
		}
	}
	return b
}

func (receiver *HpService) Register(data *message.HpMessage, conn *net2.MuxSession) {
	mu.Lock()         // 上锁
	defer mu.Unlock() // 解锁

	configkey := data.MetaData.Key
	info := receiver.loadUserConfigInfo(configkey)
	if info == nil {
		return
	}
	ClosePortServer(info.RemotePort)
	err2, protocolType, _, _ := util.ProtocolInfo(info.LocalAddress)
	if err2 != nil {
		return
	}
	connectType := bean.Protocol(protocolType)
	newTunnelServer := tunnel.NewTunnelServer(connectType, info.RemotePort, conn, *info)
	server := newTunnelServer.StartServer()
	if !server {
		newTunnelServer.CLose()
	} else {
		log.Info("隧道启动成功")
		HP_CACHE_TUN.Store(info.RemotePort, newTunnelServer)
	}
	if info.Domain != nil {
		DOMAIN_HP_INFO.Store(*info.Domain, info)
	}
	now := time.Now()
	//更新服务端状态
	strMsg := now.Format("2006年01月02日 15时04分05秒") + " 配置启动成功"
	if !server {
		strMsg = now.Format("2006年01月02日 15时04分05秒") + " 配置启动失败，大概率是端口冲突，请刷新"
	}
	db.DB.Model(&entity.UserConfigEntity{}).Where("config_key = ?", configkey).Update("status_msg", strMsg)
	//通知客户端结果
	arr2 := [][]string{
		{"穿透结果", strconv.FormatBool(server)},
	}

	if server && ((connectType == bean.TCP) || (connectType == bean.UNIX) || (connectType == bean.TCP_UDP)) {
		arr2 = append(arr2, []string{"内网地址", info.LocalAddress})
		arr2 = append(arr2, []string{"外网TCP", info.Ip + ":" + strconv.Itoa(info.RemotePort)})
	}

	if server && ((connectType == bean.HTTP) || connectType == bean.HTTPS) {
		arr2 = append(arr2, []string{"内网地址", info.LocalAddress})
		arr2 = append(arr2, []string{"外网HTTP", "http://" + info.Ip + ":" + strconv.Itoa(info.RemotePort)})
		if info.Domain != nil {
			arr2 = append(arr2, []string{"HTTP地址", "http://" + *info.Domain})
			arr2 = append(arr2, []string{"HTTPS地址", "https://" + *info.Domain})
		}
	}

	if server && (connectType == bean.SOCKS5) {
		arr2 = append(arr2, []string{"内网SOCKS5", info.LocalAddress})
		arr2 = append(arr2, []string{"外网SOCKS5", info.Ip + ":" + strconv.Itoa(info.RemotePort)})
	}

	if server && (connectType == bean.UDP || (connectType == bean.TCP_UDP)) {
		arr2 = append(arr2, []string{"内网UDP", info.LocalAddress})
		arr2 = append(arr2, []string{"外网UDP", info.Ip + ":" + strconv.Itoa(info.RemotePort)})
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

func ClosePortServer(port int) {
	log.Info("关闭外网端口服务：" + strconv.Itoa(port))
	tunnelServer, ok := HP_CACHE_TUN.Load(port)
	if ok {
		s := tunnelServer.(*tunnel.TunnelServer)
		s.CLose()
		HP_CACHE_TUN.Delete(port)
		info := s.UserInfo()
		if info.Domain != nil {
			DOMAIN_HP_INFO.Delete(*info.Domain)
		}
	}
}
