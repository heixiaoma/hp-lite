package entity

import "time"

type UserCustomEntity struct {
	Id *int `json:"id" gorm:"primaryKey;autoIncrement"`

	Username string `json:"username"`

	Password string `json:"password"`

	Desc string `json:"desc"`

	CreateTime time.Time `json:"createTime"`
}

func (UserCustomEntity) TableName() string {
	return "user_custom"
}
