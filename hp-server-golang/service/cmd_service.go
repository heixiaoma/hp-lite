package service

import (
	"encoding/json"
	"hp-server-lib/bean"
	"hp-server-lib/db"
	"hp-server-lib/entity"
	"hp-server-lib/log"
	cmdMessage "hp-server-lib/message"
	"hp-server-lib/protol"
	"net"
	"sync"
)

// string,conn
var CMD_CACHE_CONN = sync.Map{}

// string memory_info
var CMD_CACHE_MEMORY_INFO = sync.Map{}

type CmdService struct {
}

func (receiver *CmdService) sendMessage(conn net.Conn, message *cmdMessage.CmdMessage) {
	conn.Write(protol.CmdEncode(message))
}

func (receiver *CmdService) sendTips(conn net.Conn, tips string) {
	c := &cmdMessage.CmdMessage{
		Data: tips,
		Type: cmdMessage.CmdMessage_TIPS,
	}
	receiver.sendMessage(conn, c)
}

func SendCloseMsg(deviceKey, msg string) bool {
	value, ok := CMD_CACHE_CONN.Load(deviceKey)
	if ok {
		c := &cmdMessage.CmdMessage{
			Data: msg,
			Type: cmdMessage.CmdMessage_DISCONNECT,
		}
		conn := value.(net.Conn)
		conn.Write(protol.CmdEncode(c))
		conn.Close()
		CMD_CACHE_CONN.Delete(deviceKey)
	}
	return ok
}

func NoticeClientUpdateData(deviceKey string) bool {
	var results []entity.UserConfigEntity
	db.DB.Model(&entity.UserConfigEntity{}).Where("device_key = ? and (status = 0 or status is null) ", deviceKey).Find(&results)

	//读取防火墙安全规则
	var configIds []int
	for _, item := range results {
		configIds = append(configIds, *item.Id)
	}
	var configItems []*entity.UserWafEntity
	configMap := make(map[int]*entity.UserWafEntity)
	if err := db.DB.Model(&entity.UserWafEntity{}).Where("config_id IN ?", configIds).Find(&configItems).Error; err == nil {
		// 将查询结果转换成 map[int]User
		for _, conf := range configItems {
			configMap[conf.ConfigId] = conf
		}
	}

	if results != nil {
		var results2 []bean.LocalInnerWear
		for _, item := range results {
			//设置防火墙安全规则下发客户端
			OutLimit := -1
			InLimit := -1
			customEntity := configMap[*item.Id]
			if customEntity != nil {
				if customEntity.OutLimit > 0 {
					OutLimit = customEntity.OutLimit
				}
				if customEntity.InLimit > 0 {
					InLimit = customEntity.InLimit
				}
			}
			wear := bean.LocalInnerWear{
				OutLimit:     OutLimit,
				InLimit:      InLimit,
				TunType:      item.TunType,
				RemotePort:   *item.RemotePort,
				LocalAddress: item.LocalAddress,
				ConfigKey:    item.ConfigKey,
				ServerIp:     item.ServerIp,
				ServerPort:   *item.ServerPort,
			}
			results2 = append(results2, wear)
		}
		jsonData, err := json.Marshal(results2)
		if err == nil {
			c := &cmdMessage.CmdMessage{
				Data: string(jsonData),
				Type: cmdMessage.CmdMessage_LOCAL_INNER_WEAR,
			}
			value, ok := CMD_CACHE_CONN.Load(deviceKey)
			if ok {
				conn := value.(net.Conn)
				conn.Write(protol.CmdEncode(c))
			}
			return ok
		}
	}

	return false
}

func (receiver CmdService) StoreMemInfo(conn net.Conn, message *cmdMessage.CmdMessage) {
	data := message.GetData()
	if len(data) > 0 {
		info := &bean.MemoryInfo{}
		err := json.Unmarshal([]byte(data), info)
		if err == nil {
			CMD_CACHE_MEMORY_INFO.Delete(message.GetKey())
			CMD_CACHE_MEMORY_INFO.Store(message.GetKey(), info)
			CMD_CACHE_CONN.Delete(message.GetKey())
			CMD_CACHE_CONN.Store(message.GetKey(), conn)
		}
	} else {
		info := &bean.MemoryInfo{}
		CMD_CACHE_MEMORY_INFO.Delete(message.GetKey())
		CMD_CACHE_MEMORY_INFO.Store(message.GetKey(), info)
		CMD_CACHE_CONN.Delete(message.GetKey())
		CMD_CACHE_CONN.Store(message.GetKey(), conn)
	}
}

func (receiver *CmdService) Connect(conn net.Conn, message *cmdMessage.CmdMessage) {
	_, ok := CMD_CACHE_CONN.Load(message.GetKey())
	if ok {
		log.Errorf("设备KEY已经在线:%s", message.GetKey())
		receiver.sendTips(conn, "设备KEY已经在线")
		return
	} else {
		receiver.StoreMemInfo(conn, message)
		NoticeClientUpdateData(message.GetKey())
	}
}

func (receiver CmdService) Clear(conn net.Conn) {
	deviceKey := ""
	CMD_CACHE_CONN.Range(func(key, value interface{}) bool {
		deviceKey = key.(string)
		if value == conn {
			return false
		}
		// 返回 true 继续遍历
		return true
	})
	//清除数据
	if len(deviceKey) > 0 {
		log.Infof("清除设备key:%s", deviceKey)
		CMD_CACHE_CONN.Delete(deviceKey)
		CMD_CACHE_MEMORY_INFO.Delete(deviceKey)
	}
}
