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
		_, ok := localRouter.Load(key1)
		//云端找不到的数据，就需要删除掉
		if !ok {
			//todo 开启新的
			startTunnel(value1.(*bean.LocalInnerWear), callMsg)
			localRouter.Store(key1, value1)
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
	//开始进行真正的映射了
	hpClient := NewHpClient(callMsg)
	tunnel.Store(data.ConfigKey, hpClient)
	hpClient.Connect(data)
	go func() {
		flagStatus := false
		for {
			//检查本地存储是否存在，如果不存在就需要关闭了
			_, ok := localRouter.Load(data.ConfigKey)
			if ok {
				status := hpClient.GetStatus()
				if !status {
					//开始重新连接
					hpClient.CallMsg("正在重新连接")
					hpClient.Connect(data)
				}
				if flagStatus != status {
					//连接有变化就打印一次
					PrintTable(hpClient.CallMsg)
					flagStatus = status
				}
				time.Sleep(time.Duration(10) * time.Second)
			} else {
				//云端数据都删除了，我们就需要停止服务了
				hpClient.Close()
				tunnel.Delete(data.ConfigKey)
				return
			}
		}
	}()
}
