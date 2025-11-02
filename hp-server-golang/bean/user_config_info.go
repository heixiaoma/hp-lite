package bean

import (
	"net/http/httputil"
)

type UserConfigInfo struct {
	ProxyVersion string                 `json:"proxyVersion"`
	LocalAddress string                 `json:"localAddress"`
	Domain       *string                `json:"domain"`
	ConfigId     int                    `json:"configId"`
	RemotePort   int                    `json:"remotePort"`
	Ip           string                 `json:"ip"`
	TunType      string                 `json:"TunType"`
	MaxConn      int                    `json:"maxConn"`
	BlockedIps   []string               `json:"blockedIps"`
	AllowedIps   []string               `json:"allowedIps"`
	ReverseProxy *httputil.ReverseProxy `json:"-"`
}
