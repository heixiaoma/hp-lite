package bean

type LocalInnerWear struct {
	/**
	 * 映射服务器IP
	 */
	ServerIp string `json:"serverIp"`

	/**
	 * 映射服务器的端口
	 */
	ServerPort int `json:"serverPort"`

	/**
	 * 远端端口
	 */
	RemotePort int `json:"remotePort"`

	/**
	 * 本地地址
	 */
	LocalAddress string `json:"localAddress"`

	/**
	 * 本地映射的key
	 */
	ConfigKey string `json:"configKey"`

	/**
	 * 链接方式
	 */
	TunType string `json:"tunType"`

	/**
	 * 内网速度
	 */
	InLimit int `json:"inLimit"`

	/**
	 * 外网速度
	 */
	OutLimit int `json:"outLimit"`
}
