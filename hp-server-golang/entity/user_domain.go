package entity

type UserDomainEntity struct {
	/**
	 * 主键
	 */
	Id *int `json:"id" gorm:"primaryKey;autoIncrement"`

	/**
	 * 用户ID
	 */
	UserId *int `json:"userId"`

	/**
	 * 域名
	 */
	Domain *string `json:"domain"`

	/**
	 * 备注
	 */
	Desc *string `json:"desc"`

	/**
	 * SSL证书Key
	 */
	CertificateKey string `json:"certificateKey"`

	/**
	 * 证书内容
	 */
	CertificateContent string `json:"certificateContent"`

	/**
	 * 状态
	 */
	Status string `json:"status"`

	/**
	 * 提示
	 */
	Tips string `json:"tips"`
}

func (UserDomainEntity) TableName() string {
	return "user_domain"
}
