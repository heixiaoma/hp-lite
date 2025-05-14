package base

import (
	"hp-server-lib/db"
	"hp-server-lib/entity"
	"log"
	"sync"
	"time"
)

// 定义存储发送、接收、pv、uv的数据结构
type DataStats struct {
	sent     int64               // 发送的数据量
	received int64               // 接收的数据量
	pv       int64               // 页面访问量
	uv       map[string]struct{} // 用来存储独立访客的IP地址集合
	mu       sync.Mutex          // 用于保护并发访问
}

// 定义一个全局的 map 存储 configID 对应的 DataStats
var statsMap sync.Map

// addSent 方法用于更新发送的数据大小
func AddSent(configID int, sentSize int64) {
	// 从 map 中取出对应的 DataStats
	val, _ := statsMap.LoadOrStore(configID, &DataStats{uv: make(map[string]struct{})})

	// 强制类型转换
	stats := val.(*DataStats)
	stats.mu.Lock()
	stats.sent += sentSize
	stats.mu.Unlock()
}

// addReceived 方法用于更新接收的数据大小
func AddReceived(configID int, receivedSize int64) {
	// 从 map 中取出对应的 DataStats
	val, _ := statsMap.LoadOrStore(configID, &DataStats{uv: make(map[string]struct{})})

	// 强制类型转换
	stats := val.(*DataStats)
	stats.mu.Lock()
	stats.received += receivedSize
	stats.mu.Unlock()
}

// addPv 方法用于更新页面访问量 (pv)
func AddPv(configID int, pvCount int64) {
	// 从 map 中取出对应的 DataStats
	val, _ := statsMap.LoadOrStore(configID, &DataStats{uv: make(map[string]struct{})})

	// 强制类型转换
	stats := val.(*DataStats)
	stats.mu.Lock()
	stats.pv += pvCount
	stats.mu.Unlock()
}

// addUv 方法用于更新独立访客数 (uv) — 基于 IP
func AddUv(configID int, ip string) {
	// 从 map 中取出对应的 DataStats
	val, _ := statsMap.LoadOrStore(configID, &DataStats{uv: make(map[string]struct{})})

	// 强制类型转换
	stats := val.(*DataStats)
	stats.mu.Lock()
	stats.uv[ip] = struct{}{} // 使用空结构体来存储独立访客IP
	stats.mu.Unlock()
}

// 存储当前 statsMap 中每个 configID 的统计数据
func saveStats() {
	var statsData []*entity.UserStatisticsEntity
	milli := time.Now().UnixMilli()
	statsMap.Range(func(key, value interface{}) bool {
		configID := key.(int)
		stats := value.(*DataStats)
		stats.mu.Lock()
		statsData = append(statsData, &entity.UserStatisticsEntity{
			ConfigId:   configID,
			Download:   stats.sent,
			Upload:     stats.received,
			Uv:         int64(len(stats.uv)),
			Pv:         stats.pv,
			Time:       milli,
			CreateTime: time.Now(),
		})
		log.Printf("ConfigID: %d, Sent: %d bytes, Received: %d bytes, PV: %d, UV: %d\n",
			configID, stats.sent, stats.received, stats.pv, len(stats.uv))
		stats.mu.Unlock()
		return true
	})
	db.DB.CreateInBatches(statsData, len(statsData))
	// 获取 2 天前的时间
	twoDaysAgo := time.Now().AddDate(0, 0, -2)
	// 删除 createTime 小于 2 天前的数据
	db.DB.Where("create_time < ?", twoDaysAgo).Delete(&entity.UserStatisticsEntity{})
}

// 清理 map 中的数据
func clearStats() {
	saveStats()
	statsMap.Range(func(key, value interface{}) bool {
		// 清理对应的 configID 数据
		statsMap.Delete(key)
		return true
	})
	log.Println("统计数据已清除")
}

func init() {
	log.Printf("数据统计服务已启动")
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				clearStats()
			}
		}
	}()
}
