package service

import (
	"hp-server-lib/bean"
	"hp-server-lib/db"
	"hp-server-lib/entity"
	"hp-server-lib/ext"
	"hp-server-lib/ext/forward"
	"sync"
)

var FORWARD_CACHE = sync.Map{}

type ForwardService struct {
}

func (receiver *ForwardService) AddData(custom entity.UserFwdEntity) error {

	if custom.Id != nil {
		value, ok := FORWARD_CACHE.Load(*custom.Id)
		if ok {
			proxy := value.(forward.ForwardProxy)
			proxy.Stop()
		}
	}

	tx := db.DB.Save(&custom)
	if tx.Error != nil {
		return tx.Error
	}

	if *custom.Type == "1" && *custom.Status == "1" {
		server := ext.NewHttpFwdServer(*custom.Port, *custom.User, *custom.Pwd)
		start := server.Start(func() {
			FORWARD_CACHE.Delete(*custom.Id)
		})
		if start {
			FORWARD_CACHE.Store(*custom.Id, server)
		}
	}
	if *custom.Type == "2" && *custom.Status == "1" {
		server := ext.NewSocks(*custom.Port, *custom.User, *custom.Pwd)
		start := server.Start(func() {
			FORWARD_CACHE.Delete(*custom.Id)
		})
		if start {
			FORWARD_CACHE.Store(*custom.Id, server)
		}
	}
	return nil
}

func (receiver *ForwardService) ListData(userId int, page int, pageSize int) *bean.ResPage {
	var results []*entity.UserFwdEntity
	var total int64
	// 计算总记录数并执行分页查询
	if userId < 0 {
		db.DB.Model(&entity.UserFwdEntity{}).Order("id desc").Count(&total).Offset((page - 1) * pageSize).Limit(pageSize).Find(&results)
	} else {
		db.DB.Model(&entity.UserFwdEntity{}).Where("user_id = ?", userId).Order("id desc").Count(&total).Offset((page - 1) * pageSize).Limit(pageSize).Find(&results)
	}
	for _, item := range results {
		_, ok := FORWARD_CACHE.Load(*item.Id)
		if ok {
			item.Tips = "正常"
		} else {
			item.Tips = "已停止"
		}
	}
	if userId < 0 {
		var userIds []int
		for _, item := range results {
			userIds = append(userIds, *item.UserId)
		}
		var users []*entity.UserCustomEntity
		if err := db.DB.Model(&entity.UserCustomEntity{}).Where("id IN ?", userIds).Find(&users).Error; err == nil {
			// 将查询结果转换成 map[int]User
			userMap := make(map[int]*entity.UserCustomEntity)
			for _, user := range users {
				userMap[*user.Id] = user
			}
			for _, item := range results {
				customEntity := userMap[*item.UserId]
				if customEntity != nil {
					item.Username = customEntity.Username
					item.UserDesc = customEntity.Desc
				}

			}
		}
	}
	return bean.PageOk(total, results)
}

func (receiver *ForwardService) RemoveData(id int) {
	db.DB.Delete(&entity.UserFwdEntity{Id: &id})
	value, ok := FORWARD_CACHE.Load(id)
	if ok {
		proxy := value.(forward.ForwardProxy)
		proxy.Stop()
	}
}
