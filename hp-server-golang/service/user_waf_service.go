package service

import (
	"errors"
	"hp-server-lib/bean"
	"hp-server-lib/db"
	"hp-server-lib/entity"
	"net"
)

type UserWafService struct {
}

func isValidCIDR(s string) bool {
	_, _, err := net.ParseCIDR(s)
	return err == nil
}

func (receiver *UserWafService) AddData(custom entity.UserWafEntity) error {
	//规则校验
	for _, item := range custom.AllowedIPs {
		cidr := isValidCIDR(item)
		if !cidr {
			return errors.New(item + ",不符合CIDR规则")
		}
	}

	for _, item := range custom.BlockedIPs {
		cidr := isValidCIDR(item)
		if !cidr {
			return errors.New(item + ",不符合CIDR规则")
		}
	}

	if custom.Id == nil {
		//添加就要做规则检查防止多次添加
		var total int64
		db.DB.Model(&entity.UserWafEntity{}).Where("config_id = ?", custom.ConfigId).Count(&total)
		if total > 0 {
			return errors.New("穿透的安全规则已经存在，请查找你的列表配置")
		}
	}
	db.DB.Save(&custom)
	//刷新配置
	service := ConfigService{}
	_ = service.RefData(custom.ConfigId)
	return nil
}

func (receiver *UserWafService) ListData(page int, pageSize int) *bean.ResPage {
	var results []*entity.UserWafEntity
	var total int64
	// 计算总记录数并执行分页查询
	db.DB.Model(&entity.UserWafEntity{}).Count(&total).Offset((page - 1) * pageSize).Limit(pageSize).Find(&results)

	var configIds []int
	for _, item := range results {
		configIds = append(configIds, item.ConfigId)
	}
	var configItems []*entity.UserConfigEntity
	if err := db.DB.Model(&entity.UserConfigEntity{}).Where("id IN ?", configIds).Find(&configItems).Error; err == nil {
		// 将查询结果转换成 map[int]User
		configMap := make(map[int]*entity.UserConfigEntity)
		for _, conf := range configItems {
			configMap[*conf.Id] = conf
		}
		for _, item := range results {
			customEntity := configMap[item.ConfigId]
			if customEntity != nil {
				item.ConfigDesc = customEntity.Remarks
			}
		}
	}
	return bean.PageOk(total, results)
}

func (receiver *UserWafService) RemoveData(id int) {
	db.DB.Delete(&entity.UserWafEntity{Id: &id})
}
