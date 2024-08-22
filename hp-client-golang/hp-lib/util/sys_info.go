package util

import (
	"encoding/json"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"hp-lib/bean"
	"runtime"
)

func SysInfo() string {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	info := bean.SysInfo{}
	v, err := mem.VirtualMemory()
	if err == nil {
		info.TotalMem = v.Total
		info.UseMem = v.Used
	}
	start, err := cpu.Times(false)
	if err == nil {
		var totalUsed float64
		// 计算每个 CPU 核心的使用率并相加
		for i := range start {
			totalUsed += (start[i].Total() - start[i].Idle) / start[i].Total()
		}
		info.CpuRate = (totalUsed / float64(len(start))) * 100.0
	}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// 程序占用的总内存
	totalMemory := m.Sys
	// 程序当前使用的内存
	usedMemory := m.Alloc

	info.HpTotalMem = float64(totalMemory)
	info.HpUseMem = float64(usedMemory)
	marshal, _ := json.Marshal(info)
	return string(marshal)
}
