package bean

type SysInfo struct {
	TotalMem   uint64  `json:"total"`
	UseMem     uint64  `json:"useMem"`
	CpuRate    float64 `json:"cpuRate"`
	HpTotalMem float64 `json:"hpTotalMem"`
	HpUseMem   float64 `json:"hpUseMem"`
}
