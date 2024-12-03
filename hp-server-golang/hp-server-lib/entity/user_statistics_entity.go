package entity

import "time"

type UserStatisticsEntity struct {
	Id *int `json:"id" gorm:"primaryKey;autoIncrement"`

	/**
	 * 套餐ID
	 */
	ConfigId int `json:"configId"`

	/**
	 * 下载量
	 */
	Download int64 `json:"download"`

	/**
	 * 上传量
	 */
	Upload int64 `json:"upload"`

	/**
	 * uv
	 */
	Uv int64 `json:"uv"`

	/**
	 * pv
	 */
	Pv int64 `json:"pv"`

	/**
	 * 时间
	 */
	Time int64 `json:"time"`

	/**
	 * 创建时间
	 */
	CreateTime time.Time `json:"createTime"`
}

func (UserStatisticsEntity) TableName() string {
	return "user_statistics"
}
