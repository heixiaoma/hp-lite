package service

import (
	"errors"
	"hp-server-lib/bean"
	"hp-server-lib/db"
	"hp-server-lib/entity"
)

type DeviceService struct {
}

func (receiver *DeviceService) AddData(userId int, device bean.ReqDeviceInfo) error {
	if len(device.Desc) == 0 {
		return errors.New("设备备注不能为空")
	}
	if len(device.DeviceId) != 32 {
		return errors.New("设备编码只能是32字母数字组成")
	}
	var total int64
	db.DB.Model(&entity.UserDeviceEntity{}).Where("device_Key = ?", device.DeviceId).Count(&total)
	if total > 0 {
		return errors.New("设备编码已经存在")
	}

	db.DB.Save(&entity.UserDeviceEntity{
		Remarks:   device.Desc,
		UserId:    &userId,
		DeviceKey: device.DeviceId,
	})
	return nil
}

func (receiver *DeviceService) UpdateData(device bean.ReqDeviceInfo) error {
	if len(device.Desc) == 0 {
		return errors.New("设备备注不能为空")
	}
	if len(device.DeviceId) != 32 {
		return errors.New("设备编码只能是32字母数字组成")
	}
	var total int64
	db.DB.Model(&entity.UserDeviceEntity{}).Where("device_key = ?", device.DeviceId).Count(&total)
	if total != 1 {
		return errors.New("设备编码不存在")
	}
	db.DB.Model(&entity.UserDeviceEntity{}).Where("device_Key = ?", device.DeviceId).Update("remarks", device.Desc)
	return nil
}

func (receiver *DeviceService) ListData(userId int) []*bean.ResDeviceInfo {
	var results []entity.UserDeviceEntity
	// 计算总记录数并执行分页查询
	if userId < 0 {
		db.DB.Find(&results)
	} else {
		db.DB.Where("user_id = ?", userId).Find(&results)
	}
	var result2 []*bean.ResDeviceInfo
	var userIds []int
	for _, item := range results {
		value, ok := CMD_CACHE_MEMORY_INFO.Load(item.DeviceKey)
		info := bean.NewResDeviceInfo(item.DeviceKey, item.Remarks, ok)
		if ok {
			info.MemoryInfo = value.(*bean.MemoryInfo)
		}
		info.UserId = *item.UserId
		result2 = append(result2, info)
		userIds = append(userIds, *item.UserId)
	}

	if userId < 0 {
		var users []*entity.UserCustomEntity
		if err := db.DB.Model(&entity.UserCustomEntity{}).Where("id IN ?", userIds).Find(&users).Error; err == nil {
			// 将查询结果转换成 map[int]User
			userMap := make(map[int]*entity.UserCustomEntity)
			for _, user := range users {
				userMap[*user.Id] = user
			}
			for _, item := range result2 {
				customEntity := userMap[item.UserId]
				if customEntity != nil {
					item.Username = customEntity.Username
					item.UserDesc = customEntity.Desc
				}
			}
		}
	}
	return result2
}

func (receiver *DeviceService) RemoveData(deviceKey string) error {
	//检查是否存在配置
	var configTotal int64
	db.DB.Model(&entity.UserConfigEntity{}).Where("device_key = ?", deviceKey).Count(&configTotal)
	if configTotal > 0 {
		return errors.New("设备被占用，请删除映射后再来")
	}
	db.DB.Where("device_Key = ?", deviceKey).Delete(&entity.UserDeviceEntity{})
	return nil
}

func (receiver *DeviceService) StopData(deviceKey string) bool {
	return SendCloseMsg(deviceKey, "强制停止")
}
