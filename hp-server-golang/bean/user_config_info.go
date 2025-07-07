package bean

import (
	"net/http/httputil"
)

type UserConfigInfo struct {
	ProxyVersion       string                 `json:"proxyVersion"`
	ProxyIp            string                 `json:"proxyIp"`
	ProxyPort          int                    `json:"proxyPort"`
	Domain             *string                `json:"domain"`
	ConfigId           int                    `json:"configId"`
	Port               int                    `json:"port"`
	Ip                 string                 `json:"ip"`
	CertificateKey     string                 `json:"certificateKey"`
	CertificateContent string                 `json:"certificateContent"`
	WebType            string                 `json:"webType"`
	TunType            string                 `json:"TunType"`
	MaxConn            int                    `json:"maxConn"`
	BlockedIps         []string               `json:"blockedIps"`
	AllowedIps         []string               `json:"allowedIps"`
	ReverseProxy       *httputil.ReverseProxy `json:"-"`
}
