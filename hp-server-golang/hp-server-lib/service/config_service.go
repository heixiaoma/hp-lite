package service

import (
	"errors"
	"github.com/google/uuid"
	"hp-server-lib/bean"
	"hp-server-lib/config"
	"hp-server-lib/db"
	"hp-server-lib/entity"
	"log"
)

type ConfigService struct {
}

func (receiver *ConfigService) DeviceKey(userId int) []*bean.ResUserDeviceInfo {
	var results []entity.UserDeviceEntity
	if userId < 0 {
		db.DB.Find(&results)
	} else {
		db.DB.Where("user_id = ?", userId).Find(&results)
	}
	var data []*bean.ResUserDeviceInfo
	for _, item := range results {
		_, ok := CMD_CACHE_MEMORY_INFO.Load(item.DeviceKey)
		desc := ""
		if ok {
			desc = "在线-" + item.Remarks
		} else {
			desc = "离线-" + item.Remarks
		}
		info := &bean.ResUserDeviceInfo{
			UserId: *item.UserId,
			Key:    item.DeviceKey,
			Desc:   desc,
		}
		data = append(data, info)
	}
	if userId < 0 {
		var userIds []int
		for _, item := range data {
			userIds = append(userIds, item.UserId)
		}
		var users []*entity.UserCustomEntity
		if err := db.DB.Model(&entity.UserCustomEntity{}).Where("id IN ?", userIds).Find(&users).Error; err == nil {
			// 将查询结果转换成 map[int]User
			userMap := make(map[int]*entity.UserCustomEntity)
			for _, user := range users {
				userMap[*user.Id] = user
			}
			for _, item := range data {
				customEntity := userMap[item.UserId]
				if customEntity != nil {
					log.Printf(customEntity.Username)
					item.Username = customEntity.Username
					item.UserDesc = customEntity.Desc
				}
			}
		}
	}
	return data

}

func (receiver *ConfigService) ConfigList(userId int, page int, pageSize int) *bean.ResPage {
	var results []entity.UserConfigEntity
	var total int64
	if userId < 0 {
		db.DB.Model(&entity.UserConfigEntity{}).Order("id desc").Count(&total).Offset((page - 1) * pageSize).Limit(pageSize).Find(&results)
	} else {
		db.DB.Model(&entity.UserConfigEntity{}).Where("user_id = ?", userId).Order("id desc").Count(&total).Offset((page - 1) * pageSize).Limit(pageSize).Find(&results)
	}
	// 计算总记录数并执行分页查询
	return bean.PageOk(total, results)
}

func (receiver *ConfigService) RemoveData(configId int) bool {
	userQuery := &entity.UserConfigEntity{}
	db.DB.Where("id = ? ", configId).First(userQuery)
	if userQuery != nil {
		var results entity.UserConfigEntity
		db.DB.Where("id = ?", configId).Delete(&results)
		NoticeClientUpdateData(userQuery.DeviceKey)
		return true
	}
	return false
}
func (receiver *ConfigService) AddData(configEntity entity.UserConfigEntity) error {
	if len(configEntity.DeviceKey) == 0 {
		return errors.New("设备ID未选择")
	}
	if len(configEntity.Remarks) == 0 || len(configEntity.Remarks) > 50 {
		return errors.New("备注不能为空，同时不能超过50个字")
	}

	if configEntity.ConnectType == nil {
		return errors.New("穿透协议未选择")
	}

	if configEntity.Port == nil {
		return errors.New("外网端口未填写")
	}

	if configEntity.Id == nil {
		var total int64
		db.DB.Model(&entity.UserConfigEntity{}).Where("port = ?", configEntity.Port).Count(&total)
		if total > 0 {
			return errors.New("外网端口已被其他配置占用")
		}
		if configEntity.Domain != nil && len(*configEntity.Domain) > 0 {
			total = 0
			db.DB.Model(&entity.UserConfigEntity{}).Where("domain = ?", configEntity.Domain).Count(&total)
			if total > 0 {
				return errors.New("域名被使用了，请换一个")
			}
		}
	}

	newUUID, err := uuid.NewUUID()
	if err != nil {
		return errors.New(err.Error())
	}
	configEntity.ConfigKey = newUUID.String()
	deviceQuery := &entity.UserDeviceEntity{}
	db.DB.Where("device_key = ? ", configEntity.DeviceKey).First(deviceQuery)
	if deviceQuery == nil || deviceQuery.UserId == nil {
		return errors.New("设备不存在")
	}
	configEntity.UserId = deviceQuery.UserId
	configEntity.ServerIp = config.ConfigData.Tunnel.IP
	configEntity.ServerPort = &config.ConfigData.Tunnel.Port
	db.DB.Save(&configEntity)
	NoticeClientUpdateData(configEntity.DeviceKey)
	return nil
}

func (receiver *ConfigService) RefData(configId int) error {
	newUUID, err := uuid.NewUUID()
	if err != nil {
		return errors.New(err.Error())
	}
	userQuery := entity.UserConfigEntity{}
	db.DB.Where("id = ? ", configId).First(&userQuery)
	if userQuery.Id != nil {
		db.DB.Model(&entity.UserConfigEntity{}).Where("id = ?", configId).UpdateColumn("status_msg", nil).UpdateColumn("config_key", newUUID.String())
		NoticeClientUpdateData(userQuery.DeviceKey)
		return nil
	} else {
		return errors.New("更新失败")
	}
}
