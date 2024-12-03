package entity

type UserDeviceEntity struct {
	id int

	/**
	 * 设备key
	 */
	deviceKey string

	/**
	 * 描述
	 */
	remarks string

	/**
	 * 用户ID
	 */
	userId int
}

func (UserDeviceEntity) TableName() string {
	return "user_device"
}
