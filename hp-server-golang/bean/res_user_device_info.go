package bean

type ResUserDeviceInfo struct {
	Key      string `json:"key"`
	Desc     string `json:"desc"`
	UserId   int    `json:"userId"`
	Username string `json:"username"`
	UserDesc string `json:"userDesc"`
}
