package entity

type UserDeviceEntity struct {
	Id *int `json:"id" gorm:"primaryKey;autoIncrement"`

	/**
	 * 设备key
	 */
	DeviceKey string `json:"deviceKey"`

	/**
	 * 描述
	 */
	Remarks string `json:"remarks"`

	/**
	 * 用户ID
	 */
	UserId *int `json:"userId"`
}

func (UserDeviceEntity) TableName() string {
	return "user_device"
}
