package bean

type UserConfigInfo struct {
	ProxyVersion       string  `json:"proxyVersion"`
	ProxyIp            string  `json:"proxyIp"`
	ProxyPort          int     `json:"proxyPort"`
	Domain             *string `json:"domain"`
	ConfigId           int     `json:"configId"`
	Port               int     `json:"port"`
	Ip                 string  `json:"ip"`
	CertificateKey     string  `json:"certificateKey"`
	CertificateContent string  `json:"certificateContent"`
}
