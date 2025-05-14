package bean

type MemoryInfo struct {
	Total      int64   `json:"total"`
	UseMem     int64   `json:"useMem"`
	CpuRate    float64 `json:"cpuRate"`
	HpTotalMem int64   `json:"hpTotalMem"`
	HpUseMem   int64   `json:"hpUseMem"`
}
