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
	 * 本地IP
	 */
	LocalIp string `json:"localIp"`

	/**
	 * 本地的端口
	 */
	LocalPort int `json:"localPort"`

	/**
	 * 本地映射的key
	 */
	ConfigKey string `json:"configKey"`

	/**
	 * 映射类型
	 */
	ConnectType ConnectType `json:"connectType"`

	/**
	 * 内网速度
	 */
	InLimit int `json:"inLimit"`

	/**
	 * 外网速度
	 */
	OutLimit int `json:"outLimit"`
}
