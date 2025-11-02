package service

import (
	"errors"
	"hp-server-lib/bean"
	"hp-server-lib/config"
	"hp-server-lib/db"
	"hp-server-lib/entity"
	"hp-server-lib/util"
	"net"
	"strconv"
	"time"

	"github.com/google/uuid"
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
					item.Username = customEntity.Username
					item.UserDesc = customEntity.Desc
				}
			}
		}
	}
	return data

}

func (receiver *ConfigService) ConfigList(userId int, page int, pageSize int, keyword string) *bean.ResPage {
	var results []entity.UserConfigEntity
	var total int64

	// 基础查询
	query := db.DB.Model(&entity.UserConfigEntity{})
	if userId >= 0 {
		query = query.Where("user_id = ?", userId)
	}
	if keyword != "" {
		// 示例：匹配 username 或 config_name 字段，使用 LIKE 模糊查询
		query = query.Where("local_address LIKE ? OR remarks LIKE ? OR domain LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 分页查询
	query.Order("id desc").
		Count(&total).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&results)

	// 计算总记录数并执行分页查询
	return bean.PageOk(total, results)
}

func (receiver *ConfigService) RemoveData(configId int) bool {
	//删除防火墙的配置
	db.DB.Where("config_id = ?", configId).Delete(&entity.UserWafEntity{})
	userQuery := &entity.UserConfigEntity{}
	db.DB.Where("id = ? ", configId).First(userQuery)
	if userQuery != nil {
		var results entity.UserConfigEntity
		db.DB.Where("id = ?", configId).Delete(&results)
		NoticeClientUpdateData(userQuery.DeviceKey)
		//关闭服务端口
		ClosePortServer(*userQuery.RemotePort)
		return true
	}
	return false
}

func (receiver *ConfigService) KeywordData(userId int, keyword string) []entity.UserConfigEntity {
	var results []entity.UserConfigEntity
	key := "%" + keyword + "%"
	if userId < 0 {
		db.DB.Model(&entity.UserConfigEntity{}).Where("remarks like ? or id like ?", key, key).Limit(10).Find(&results)
	} else {
		db.DB.Model(&entity.UserConfigEntity{}).Where("user_id = ? and (remarks like ? or id like ?)", userId, key, key).Limit(10).Find(&results)
	}
	return results
}

func (receiver *ConfigService) AddData(configEntity entity.UserConfigEntity) error {
	if len(configEntity.DeviceKey) == 0 {
		return errors.New("设备ID未选择")
	}
	if len(configEntity.Remarks) == 0 || len(configEntity.Remarks) > 50 {
		return errors.New("备注不能为空，同时不能超过50个字")
	}

	if configEntity.RemotePort == nil {
		return errors.New("外网端口未填写")
	}

	//校验内网地址
	err2, _, _, _ := util.ProtocolInfo(configEntity.LocalAddress)
	if err2 != nil {
		return errors.New(err2.Error())
	}

	//检查端口占用
	// 如果是更新现有配置，需要检查端口是否发生变化
	var shouldCheckPortOccupied = true
	if configEntity.Id != nil {
		// 查询当前配置的端口
		var currentConfig entity.UserConfigEntity
		db.DB.Where("id = ?", configEntity.Id).First(&currentConfig)
		if currentConfig.RemotePort != nil && *currentConfig.RemotePort == *configEntity.RemotePort {
			// 端口没有变化，跳过端口占用检测
			shouldCheckPortOccupied = false
		} else {
			//关闭服务端口
			ClosePortServer(*currentConfig.RemotePort)
		}
	}

	if shouldCheckPortOccupied {
		conn, err := net.DialTimeout("tcp", "127.0.0.1:"+strconv.Itoa(*configEntity.RemotePort), 3*time.Second)
		if conn != nil {
			conn.Close()
		}
		if err == nil {
			return errors.New("外网端口:" + strconv.Itoa(*configEntity.RemotePort) + "已占用")
		}
	}

	if configEntity.Id == nil {
		var total int64
		db.DB.Model(&entity.UserConfigEntity{}).Where("remote_port = ?", configEntity.RemotePort).Count(&total)
		if total > 0 {
			return errors.New("外网端口:" + strconv.Itoa(*configEntity.RemotePort) + "已被其他配置占用")
		}
		if configEntity.Domain != nil && len(*configEntity.Domain) > 0 {
			total = 0
			db.DB.Model(&entity.UserConfigEntity{}).Where("domain = ?", configEntity.Domain).Count(&total)
			if total > 0 {
				return errors.New("域名被使用了，请换一个")
			}
		}
	} else {
		var total int64
		db.DB.Model(&entity.UserConfigEntity{}).Where("remote_port = ? and id != ?", configEntity.RemotePort, configEntity.Id).Count(&total)
		if total > 0 {
			return errors.New("外网端口已被其他配置占用")
		}
		if configEntity.Domain != nil && len(*configEntity.Domain) > 0 {
			total = 0
			db.DB.Model(&entity.UserConfigEntity{}).Where("domain = ? and id != ?", configEntity.Domain, configEntity.Id).Count(&total)
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

func (receiver *ConfigService) ChangeStatusData(configId int) error {
	userQuery := entity.UserConfigEntity{}
	db.DB.Where("id = ? ", configId).First(&userQuery)
	changeTmp := 1
	if userQuery.Status == 1 {
		changeTmp = 0
	} else {
		changeTmp = 1
	}
	if userQuery.Id != nil {
		db.DB.Model(&entity.UserConfigEntity{}).Where("id = ?", configId).UpdateColumn("status", changeTmp)
		return receiver.RefData(configId)
	}
	return nil
}
