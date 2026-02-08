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
	 * 本地地址
	 */
	LocalAddress string `json:"localAddress"`

	/**
	 * 备注
	 */
	Remarks string `json:"remarks"`

	/**
	 * 外网端口端口
	 */
	RemotePort *int `json:"remotePort"`

	/**
	 * HTTP模式下的域名
	 */
	Domain *string `json:"domain"`

	/**
	 * 防护类型、0 无防护(域名+外网端口都可访问)、1 全防护（域名防护+外网端口不可访问） 、2 半防护（域名防护+外网端口可访问-无防护）
	 */
	SafeType int `json:"safeType"`
	/**
	 * 防护配置ID
	 */
	SafeId int `json:"safeId"`

	/**
	 * 状态
	 */
	StatusMsg string `json:"statusMsg"`

	/**
	 * 代理版本
	 */
	ProxyVersion bean.ProxyVersion `json:"proxyVersion"`

	/**
	 * 用户ID
	 */
	UserId *int `json:"userId"`

	/**
	 * 隧道类型QUIC/TCP
	 */
	TunType string `json:"tunType"`
	/**
	 * 状态 0生效 1 不生效
	 */
	Status int `json:"status"`
}

func (UserConfigEntity) TableName() string {
	return "user_config"
}
