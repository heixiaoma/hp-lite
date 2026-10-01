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
	"time"
)

// string,conn -> *cmdConn
var CMD_CACHE_CONN = sync.Map{}

// string memory_info
var CMD_CACHE_MEMORY_INFO = sync.Map{}

// CMD_REG_LOCK 串行化 注册/踢线/清理 三个动作，保证"释放旧连接"和"注册新连接"是原子操作
var CMD_REG_LOCK sync.Mutex

const (
	// cmdConnIdleTimeout 连接空闲超过该时长即判定为已死亡，允许新连接接管。
	// 背景：客户端断网/休眠后 TCP 处于半开状态，服务端收不到 FIN，KEY 会被一直占用，
	// 客户端重连时就会被 "设备KEY已经在线" 拒之门外。客户端心跳约 10s，这里取 45s 留足冗余。
	cmdConnIdleTimeout = 45 * time.Second
	// cmdConnReapInterval 服务端巡检周期
	cmdConnReapInterval = 30 * time.Second
	// cmdConnReapIdle 巡检时发现连接空闲超过该时长则主动释放，回收半开连接占用的资源
	cmdConnReapIdle = 3 * time.Minute
)

// cmdConn 设备KEY当前持有的连接
type cmdConn struct {
	Conn       net.Conn
	Key        string
	CreateTime time.Time
	// ActiveTime 最近一次收到该连接消息的时间，用于判断连接是否还活着
	ActiveTime time.Time
}

func getCmdConn(key string) (*cmdConn, bool) {
	value, ok := CMD_CACHE_CONN.Load(key)
	if !ok {
		return nil, false
	}
	c, ok := value.(*cmdConn)
	return c, ok
}

func writeTips(conn net.Conn, tips string) {
	if conn == nil {
		return
	}
	_, _ = conn.Write(protol.CmdEncode(&cmdMessage.CmdMessage{
		Data: tips,
		Type: cmdMessage.CmdMessage_TIPS,
	}))
}

// evictCmdConn 释放连接占用的KEY。调用前必须持有 CMD_REG_LOCK
func evictCmdConn(c *cmdConn, tips string) {
	if c == nil {
		return
	}
	// 只有KEY仍然指向这条连接时才清理，避免误删已经接管的新连接
	if cur, ok := getCmdConn(c.Key); ok && cur.Conn == c.Conn {
		CMD_CACHE_CONN.Delete(c.Key)
		CMD_CACHE_MEMORY_INFO.Delete(c.Key)
	}
	if tips != "" {
		writeTips(c.Conn, tips)
	}
	if c.Conn != nil {
		_ = c.Conn.Close()
	}
}

// reapIdleConn 后台巡检：清理长时间没有消息的半开连接，避免设备KEY被永久占用
func reapIdleConn() {
	ticker := time.NewTicker(cmdConnReapInterval)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		CMD_REG_LOCK.Lock()
		var dead []*cmdConn
		CMD_CACHE_CONN.Range(func(key, value interface{}) bool {
			if c, ok := value.(*cmdConn); ok && now.Sub(c.ActiveTime) > cmdConnReapIdle {
				dead = append(dead, c)
			}
			return true
		})
		for _, c := range dead {
			log.Infof("设备KEY:%s 空闲超过%v,服务端主动释放连接", c.Key, cmdConnReapIdle)
			evictCmdConn(c, "")
		}
		CMD_REG_LOCK.Unlock()
	}
}

func init() {
	go reapIdleConn()
}

type CmdService struct {
}

func (receiver *CmdService) sendMessage(conn net.Conn, message *cmdMessage.CmdMessage) {
	if conn == nil {
		return
	}
	conn.Write(protol.CmdEncode(message))
}

func (receiver *CmdService) sendTips(conn net.Conn, tips string) {
	writeTips(conn, tips)
}

func SendCloseMsg(deviceKey, msg string) bool {
	CMD_REG_LOCK.Lock()
	defer CMD_REG_LOCK.Unlock()
	value, ok := CMD_CACHE_CONN.Load(deviceKey)
	if ok {
		client, ok := value.(*cmdConn)
		if !ok || client.Conn == nil {
			return false
		}
		m := &cmdMessage.CmdMessage{
			Data: msg,
			Type: cmdMessage.CmdMessage_DISCONNECT,
		}
		client.Conn.Write(protol.CmdEncode(m))
		client.Conn.Close()
		CMD_CACHE_CONN.Delete(deviceKey)
		CMD_CACHE_MEMORY_INFO.Delete(deviceKey)
		return true
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
				if client, ok := value.(*cmdConn); ok && client.Conn != nil {
					client.Conn.Write(protol.CmdEncode(c))
					return true
				}
			}
			return false
		}
	}

	return false
}

// StoreMemInfo 记录设备信息并刷新连接活跃时间。
// 只有当前持有该KEY的连接才能写入，防止已被接管的旧连接把状态写回来造成两端来回互踢。
func (receiver CmdService) StoreMemInfo(conn net.Conn, message *cmdMessage.CmdMessage) {
	CMD_REG_LOCK.Lock()
	defer CMD_REG_LOCK.Unlock()
	receiver.storeMemInfoLocked(conn, message)
}

// storeMemInfoLocked 调用前必须持有 CMD_REG_LOCK
func (receiver CmdService) storeMemInfoLocked(conn net.Conn, message *cmdMessage.CmdMessage) {
	key := message.GetKey()
	if len(key) == 0 || conn == nil {
		return
	}
	if cur, ok := getCmdConn(key); ok {
		if cur.Conn != conn {
			// 该KEY已经被其他连接接管，这条连接必须重新走 CONNECT 注册流程；
			// 直接关闭它，客户端会立即触发自动重连，由服务端决定是拒绝还是放行
			log.Infof("设备KEY:%s 已被新连接接管,释放旧连接", key)
			writeTips(conn, "设备KEY已在其他连接上线,正在重新连接")
			_ = conn.Close()
			return
		}
		// 心跳：只刷新活跃时间，用来判断连接是否还活着
		cur.ActiveTime = time.Now()
	} else {
		CMD_CACHE_CONN.Store(key, &cmdConn{Conn: conn, Key: key, CreateTime: time.Now(), ActiveTime: time.Now()})
	}
	info := &bean.MemoryInfo{}
	if data := message.GetData(); len(data) > 0 {
		if err := json.Unmarshal([]byte(data), info); err != nil {
			info = &bean.MemoryInfo{}
		}
	}
	CMD_CACHE_MEMORY_INFO.Store(key, info)
}

// Connect 客户端注册。同一个KEY重连时：
// 1. 旧连接已停止心跳（大概率是掉线后的半开连接） -> 释放旧连接，由新连接接管；
// 2. 旧连接仍在正常心跳 -> 拒绝并关闭新连接，客户端会按退避策略重试，不会永久卡死。
func (receiver *CmdService) Connect(conn net.Conn, message *cmdMessage.CmdMessage) {
	key := message.GetKey()
	if len(key) == 0 {
		log.Errorf("设备KEY为空,拒绝注册")
		receiver.sendTips(conn, "设备KEY为空")
		return
	}

	CMD_REG_LOCK.Lock()
	defer CMD_REG_LOCK.Unlock()

	if old, ok := getCmdConn(key); ok {
		if old.Conn == conn {
			// 同一条连接重复注册，幂等处理
			receiver.storeMemInfoLocked(conn, message)
			return
		}
		if idle := time.Since(old.ActiveTime); idle <= cmdConnIdleTimeout {
			// 旧连接还在正常通信，不允许抢占，否则两个客户端会互相踢形成死循环
			log.Errorf("设备KEY已经在线:%s|旧连接最近活跃:%v", key, old.ActiveTime)
			receiver.sendTips(conn, "设备KEY已经在线,请稍后重试")
			// 关闭被拒绝的连接，客户端才能感知到"没连上"并继续重连
			_ = conn.Close()
			return
		} else {
			// 旧连接长时间无任何消息：客户端掉线重连的典型表现，允许新连接接管
			log.Infof("设备KEY:%s 的旧连接已空闲%v,判定为掉线,新连接接管", key, idle)
			evictCmdConn(old, "设备KEY已在其他连接上线,本连接已被释放")
		}
	}

	receiver.storeMemInfoLocked(conn, message)
	NoticeClientUpdateData(key)
}

// Clear 连接断开时清理其占用的KEY。
// 注意：只有KEY仍指向这条连接时才清理，避免迟到的断连事件把已经接管的新连接删掉。
func (receiver CmdService) Clear(conn net.Conn) {
	if conn == nil {
		return
	}
	CMD_REG_LOCK.Lock()
	defer CMD_REG_LOCK.Unlock()

	var target *cmdConn
	CMD_CACHE_CONN.Range(func(key, value interface{}) bool {
		if c, ok := value.(*cmdConn); ok && c.Conn == conn {
			target = c
			// 返回 false 停止遍历
			return false
		}
		// 返回 true 继续遍历
		return true
	})
	if target == nil {
		return
	}
	if cur, ok := getCmdConn(target.Key); !ok || cur.Conn != conn {
		log.Infof("设备KEY:%s 已被新连接接管,忽略旧连接的断连事件", target.Key)
		return
	}
	//清除数据
	log.Infof("清除设备key:%s", target.Key)
	CMD_CACHE_CONN.Delete(target.Key)
	CMD_CACHE_MEMORY_INFO.Delete(target.Key)
}
