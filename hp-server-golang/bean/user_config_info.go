package bean

import (
	"net/http"
)

type UserConfigInfo struct {
	ProxyVersion string       `json:"proxyVersion"`
	LocalAddress string       `json:"localAddress"`
	Domain       *string      `json:"domain"`
	ConfigId     int          `json:"configId"`
	RemotePort   int          `json:"remotePort"`
	Ip           string       `json:"ip"`
	TunType      string       `json:"TunType"`
	MaxConn      int          `json:"maxConn"`
	SafeType     int          `json:"safeType"`
	SafeId       int          `json:"SafeId"`
	BlockedIps   []string     `json:"blockedIps"`
	AllowedIps   []string     `json:"allowedIps"`
	ReverseProxy http.Handler `json:"-"`
}
