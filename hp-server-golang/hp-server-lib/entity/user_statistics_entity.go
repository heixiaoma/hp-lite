package entity

import "time"

type UserStatisticsEntity struct {
	id int

	/**
	 * 套餐ID
	 */
	configId int

	/**
	 * 下载量
	 */
	download int64

	/**
	 * 上传量
	 */
	upload int64

	/**
	 * uv
	 */
	uv int64

	/**
	 * pv
	 */
	pv int64

	/**
	 * 时间
	 */
	time int64

	/**
	 * 创建时间
	 */
	createTime time.Time
}

func (UserStatisticsEntity) TableName() string {
	return "user_statistics"
}
