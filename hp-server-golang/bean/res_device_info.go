package bean

type ResDeviceInfo struct {
	UserId int `json:"userId"`

	Username string `json:"username"`

	UserDesc string `json:"userDesc"`

	DeviceId string `json:"deviceId"`

	Desc string `json:"desc"`

	Online bool `json:"online"`

	MemoryInfo *MemoryInfo `json:"memoryInfo"`
}

func NewResDeviceInfo(deviceId string, desc string, online bool) *ResDeviceInfo {
	return &ResDeviceInfo{
		DeviceId: deviceId,
		Desc:     desc,
		Online:   online,
	}
}
