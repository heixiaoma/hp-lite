package entity

import "net/http/httputil"

// 开启OpenDomain 才行
type UserReverseEntity struct {
	/**
	 * 主键
	 */
	Id *int `json:"id" gorm:"primaryKey;autoIncrement"`

	/**
	 * 用户ID
	 */
	UserId *int `json:"userId"`

	Domain *string `json:"domain"`

	Address *string `json:"address"`
	/**
	 * 备注
	 */
	Desc *string `json:"desc"`

	Username string `json:"username" gorm:"-"`

	UserDesc string `json:"userDesc"  gorm:"-"`

	ReverseProxy *httputil.ReverseProxy `json:"-" gorm:"-"`
}

func (UserReverseEntity) TableName() string {
	return "user_reverse"
}
