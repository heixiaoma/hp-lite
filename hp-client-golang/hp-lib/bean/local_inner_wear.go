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
	 * 配置的MD5值
	 */
	Md5 string `json:"md5"`

	/**
	 * 内网速度
	 */
	InLimit int `json:"inLimit"`

	/**
	 * 外网速度
	 */
	OutLimit int `json:"outLimit"`

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
