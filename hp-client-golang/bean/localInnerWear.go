package bean

import (
	"encoding/json"
	"log"
)

type ConnectType string

const (
	TCP     ConnectType = "TCP"
	UDP     ConnectType = "UDP"
	TCP_UDP ConnectType = "TCP_UDP"
)

type LocalInnerWear struct {
	/**
	 * 穿透服务器IP
	 */
	ServerIp string `json:"serverIp"`

	/**
	 * 穿透服务器的端口
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
	 * 本地穿透的key
	 */
	ConfigKey string `json:"configKey"`

	/**
	 * 穿透类型
	 */
	ConnectType ConnectType `json:"connectType"`

	/**
	 * 配置的MD5值
	 */
	Md5 string `json:"md5"`

	/**
	 * 本地使用的类型
	 */
	Status bool `json:"-"`
}

// NewLocalInnerWear /** 字符串转对象*/
func NewLocalInnerWear(msg string) []*LocalInnerWear {
	var wears []*LocalInnerWear
	err := json.Unmarshal([]byte(msg), &wears)
	if err != nil {
		log.Fatal(err)
	}
	return wears
}
