package entity

import "hp-server-lib/bean"

type UserConfigEntity struct {

	/**
	 * 主键
	 */
	id int
	/**
	 *当前key
	 */
	configKey string
	/**
	 * 用户KEY
	 */
	deviceKey string
	/**
	 * 套餐IP
	 */
	serverIp string

	/**
	 * 套餐端口
	 */

	serverPort int

	/**
	 * 本地IP
	 */
	localIp string

	/**
	 * 本地端口
	 */
	localPort int

	/**
	 * 穿透类型
	 */
	connectType bean.ConnectType

	/**
	 * 备注
	 */
	remarks string

	/**
	 * 端口
	 */
	port int

	/**
	 * 域名
	 */
	domain string

	/**
	 * 状态
	 */
	statusMsg string

	/**
	 * 代理版本
	 */
	proxyVersion bean.ProxyVersion

	/**
	 * SSL证书Key
	 */
	certificateKey string

	/**
	 * 证书内容
	 */
	certificateContent string

	/**
	 * 用户ID
	 */
	userId int
}

func (UserConfigEntity) TableName() string {
	return "user_config"
}
