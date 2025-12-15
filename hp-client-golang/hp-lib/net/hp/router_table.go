package hp

import (
	"hp-lib/bean"
	"hp-lib/util"
	"sync"
	"time"
)

// cloudRouter 云端最新的配置
var cloudRouter = sync.Map{}

// localRouter 本地最新的映射
var localRouter = sync.Map{}

// tunnel 隧道数据
var tunnel = sync.Map{}

var syncLock sync.Mutex // 全局锁

func CloseTunnel() {
	tunnel.Range(func(key, value any) bool {
		client := value.(*HpClient)
		client.Close()
		tunnel.Delete(key)
		return true
	})

	localRouter.Range(func(key, value any) bool {
		localRouter.Delete(key)
		return true
	})

	cloudRouter.Range(func(key, value any) bool {
		cloudRouter.Delete(key)
		return true
	})

}

// RefreshRouter 刷新本地的云端配置数据
func RefreshRouter(wears []*bean.LocalInnerWear, callMsg func(msg string)) {

	syncLock.Lock()
	defer syncLock.Unlock() // 确保锁最终释放

	// 遍历缓存并删除所有数据
	cloudRouter.Range(func(key, value any) bool {
		cloudRouter.Delete(key)
		return true
	})
	//存储新的云端数据
	for _, wear := range wears {
		cloudRouter.Store(wear.ConfigKey, wear)
	}
	//检查本地数据是否和云端一致，需要保证一致处理
	//本地和云端比较多的删除
	localRouter.Range(func(key1, value1 any) bool {
		_, ok := cloudRouter.Load(key1)
		//云端找不到的数据，就需要删除掉
		if !ok {
			//todo 停止本地的映射
			v, ok := tunnel.Load(key1)
			if ok {
				client := v.(*HpClient)
				client.Close()
				tunnel.Delete(key1)
			}
			localRouter.Delete(key1)
		}
		return true
	})

	//云端和本地比较少得添加
	cloudRouter.Range(func(key1, value1 any) bool {
		configKey := key1.(string) // 假设 ConfigKey 是字符串
		_, ok := localRouter.LoadOrStore(configKey, value1)
		if !ok { // 首次存储成功，才启动 tunnel
			startTunnel(value1.(*bean.LocalInnerWear), callMsg)
		} else {
			// 已存在，无需重复启动（可打印日志排查重复调用原因）
			callMsg("配置 " + configKey + " 已存在，无需重复启动隧道")
		}
		return true
	})
}

func PrintTable(callMsg func(msg string)) {
	data := make([]*bean.LocalInnerWear, 0) // 创建长度为5的动态数组
	tunnel.Range(func(key, value any) bool {
		client := value.(*HpClient)
		client.Data.Status = client.GetStatus()
		data = append(data, client.Data)
		return true
	})
	callMsg(util.PrintStatus(data))
}

// startTunnel 开启新的隧道
func startTunnel(data *bean.LocalInnerWear, callMsg func(msg string)) {
	configKey := data.ConfigKey
	//开始进行真正的映射了
	if oldClient, ok := tunnel.Load(configKey); ok {
		oldHpClient := oldClient.(*HpClient)
		close(oldHpClient.quit) // 关闭旧的退出通道，让旧 goroutine 退出
		oldHpClient.Close()     // 关闭旧连接
		tunnel.Delete(data.ConfigKey)
	}

	hpClient := NewHpClient(callMsg)
	tunnel.Store(data.ConfigKey, hpClient)
	hpClient.Connect(data)

	go func() {
		flagStatus := false
		ticker := time.NewTicker(10 * time.Second)
		defer func() {
			ticker.Stop()
			hpClient.Close()
			tunnel.Delete(configKey)
		}()

		for {
			select {
			case <-hpClient.quit:
				hpClient.CallMsg("隧道监控 goroutine 已退出:" + data.LocalAddress)
				return
			case <-ticker.C: // 定时检查
				_, ok := localRouter.Load(configKey)
				if !ok {
					return // 本地已删除，退出
				}
				status := hpClient.GetStatus()
				if !status {
					hpClient.CallMsg("隧道正在重新连接:" + data.LocalAddress)
					hpClient.Connect(data)
				}
				if flagStatus != status {
					PrintTable(hpClient.CallMsg)
					flagStatus = status
				}
			}
		}
	}()
}
