package service

import (
	"encoding/json"
	"hp-server-lib/bean"
	cmdMessage "hp-server-lib/message"
	"hp-server-lib/protol"
	"log"
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

func (receiver CmdService) storeMemInfo(message *cmdMessage.CmdMessage) {
	data := message.GetData()
	if len(data) > 0 {
		info := &bean.MemoryInfo{}
		err := json.Unmarshal([]byte(data), info)
		if err == nil {
			println(info.HpTotalMem)
			CMD_CACHE_MEMORY_INFO.Store(message.GetKey(), info)
		}
	}
}

func (receiver CmdService) connectInfo(conn net.Conn, message *cmdMessage.CmdMessage) {
	a := &bean.LocalInnerWear{
		OutLimit:    -1,
		InLimit:     -1,
		ConnectType: "TCP",
		ConfigKey:   "ac9b0b97-b3dd-44ef-9570-0a0039823398",
		LocalIp:     "192.168.100.1",
		LocalPort:   80,
		ServerIp:    "127.0.0.1",
		ServerPort:  9090,
	}
	arr2 := [1]*bean.LocalInnerWear{a}
	jsonData, err := json.Marshal(arr2)
	if err == nil {
		c := &cmdMessage.CmdMessage{
			Data: string(jsonData),
			Type: cmdMessage.CmdMessage_LOCAL_INNER_WEAR,
		}
		receiver.sendMessage(conn, c)
	}
}

func (receiver *CmdService) Connect(conn net.Conn, message *cmdMessage.CmdMessage) {
	_, ok := CMD_CACHE_CONN.Load(message.GetKey())
	if ok {
		log.Printf("设备KEY已经在线:%s", message.GetKey())
		receiver.sendTips(conn, "设备KEY已经在线")
		return
	} else {
		receiver.storeMemInfo(message)
		CMD_CACHE_CONN.Store(message.GetKey(), conn)
		receiver.connectInfo(conn, message)
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
		log.Printf("清除设备key:%s", deviceKey)
		CMD_CACHE_CONN.Delete(deviceKey)
		CMD_CACHE_MEMORY_INFO.Delete(deviceKey)
	}
}
