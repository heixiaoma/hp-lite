package entity

type UserSafeEntity struct {
	/**
	 * 主键
	 */
	Id *int `json:"id" gorm:"primaryKey;autoIncrement"`
	/**
	 * 用户ID
	 */
	UserId int `json:"userId"`

	Rule string `json:"rule"`

	RuleName string `json:"ruleName"`

	Username string `json:"username" gorm:"-"`

	UserDesc string `json:"userDesc"  gorm:"-"`
}

func (UserSafeEntity) TableName() string {
	return "user_safe"
}
