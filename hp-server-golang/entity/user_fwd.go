package entity

type UserFwdEntity struct {
	/**
	 * 主键
	 */
	Id *int `json:"id" gorm:"primaryKey;autoIncrement"`

	/**
	 * 用户ID
	 */
	UserId *int `json:"userId"`

	Port *string `json:"port"`

	User *string `json:"user"`

	Pwd *string `json:"pwd"`

	// 1 http/https  2 socks
	Type *string `json:"type"`
	//1 启用 0 关闭
	Status *string `json:"status"`

	/**
	 * 备注
	 */
	Desc *string `json:"desc"`

	Tips string `json:"tips" gorm:"-"`

	Username string `json:"username" gorm:"-"`

	UserDesc string `json:"userDesc"  gorm:"-"`
}

func (UserFwdEntity) TableName() string {
	return "user_fwd"
}
