package entity

import "time"

type EmailCodeEntity struct {
	Id *int `json:"id" gorm:"primaryKey;autoIncrement"`

	Email string `json:"email"`

	Code string `json:"code"`

	Type string `json:"type"` // 验证码类型：verify_email, reset_password

	UserId *int `json:"userId"` // 关联的用户ID，如果是找回密码则有值

	Used bool `json:"used"` // 是否已使用

	CreateTime time.Time `json:"createTime"`

	ExpireTime time.Time `json:"expireTime"` // 过期时间
}

func (EmailCodeEntity) TableName() string {
	return "email_code"
}
