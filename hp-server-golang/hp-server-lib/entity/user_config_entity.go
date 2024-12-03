package entity

import "hp-server-lib/bean"

type UserConfigEntity struct {
	/**
	 * 主键
	 */
	Id *int `json:"id" gorm:"primaryKey;autoIncrement"`
	/**
	 *当前key
	 */
	ConfigKey string `json:"configKey"`
	/**
	 * 用户KEY
	 */
	DeviceKey string `json:"deviceKey"`
	/**
	 * 套餐IP
	 */
	ServerIp string `json:"serverIp"`

	/**
	 * 套餐端口
	 */

	ServerPort *int `json:"serverPort"`

	/**
	 * 本地IP
	 */
	LocalIp string `json:"localIp"`

	/**
	 * 本地端口
	 */
	LocalPort *int `json:"localPort"`

	/**
	 * 穿透类型
	 */
	ConnectType *bean.ConnectType `json:"connectType"`

	/**
	 * 备注
	 */
	Remarks string `json:"remarks"`

	/**
	 * 端口
	 */
	Port *int `json:"port"`

	/**
	 * 域名
	 */
	Domain *string `json:"domain"`

	/**
	 * 状态
	 */
	StatusMsg string `json:"statusMsg"`

	/**
	 * 代理版本
	 */
	ProxyVersion bean.ProxyVersion `json:"proxyVersion"`

	/**
	 * SSL证书Key
	 */
	CertificateKey string `json:"certificateKey"`

	/**
	 * 证书内容
	 */
	CertificateContent string `json:"certificateContent"`

	/**
	 * 用户ID
	 */
	UserId *int `json:"userId"`
}

func (UserConfigEntity) TableName() string {
	return "user_config"
}
