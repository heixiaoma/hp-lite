package service

import (
	"hp-server-lib/bean"
	"hp-server-lib/db"
	"hp-server-lib/entity"
	"sync"
)

var DOMAIN_REVERSE_INFO = sync.Map{}

type ReverseService struct {
}

func InitReverseECache() {
	page := 1
	pageSize := 100
	for {
		var results []*entity.UserReverseEntity
		tx := db.DB.Model(&entity.UserReverseEntity{}).
			Offset((page - 1) * pageSize).
			Limit(pageSize).
			Find(&results)
		if tx.Error != nil {
			break
		}
		// 如果本页没有数据，说明结束
		if len(results) == 0 {
			break
		}
		// 放入缓存
		for _, r := range results {
			DOMAIN_REVERSE_INFO.Store(*r.Domain, r)
		}
		// 下一页
		page++
	}
}

func (receiver *ReverseService) AddData(custom entity.UserReverseEntity) error {
	tx := db.DB.Save(&custom)
	// 创建一个新的变量副本，避免存储局部变量指针
	// 这里使用指针接收数据库返回的结果
	saved := custom
	// 明确存储指针类型
	DOMAIN_REVERSE_INFO.Store(*saved.Domain, &saved)
	return tx.Error
}

func (receiver *ReverseService) ListData(userId int, page int, pageSize int) *bean.ResPage {
	var results []*entity.UserReverseEntity
	var total int64
	// 计算总记录数并执行分页查询
	if userId < 0 {
		db.DB.Model(&entity.UserReverseEntity{}).Order("id desc").Count(&total).Offset((page - 1) * pageSize).Limit(pageSize).Find(&results)
	} else {
		db.DB.Model(&entity.UserReverseEntity{}).Where("user_id = ?", userId).Order("id desc").Count(&total).Offset((page - 1) * pageSize).Limit(pageSize).Find(&results)
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

func (receiver *ReverseService) RemoveData(id int) {
	userQuery := &entity.UserReverseEntity{}
	db.DB.Where("id = ? ", id).First(userQuery)
	if userQuery != nil {
		DOMAIN_REVERSE_INFO.Delete(*userQuery.Domain)
	}
	db.DB.Delete(&entity.UserReverseEntity{Id: &id})
}
