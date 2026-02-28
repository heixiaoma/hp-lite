package service

import (
	"github.com/corazawaf/coraza/v3"
	"hp-server-lib/bean"
	"hp-server-lib/db"
	"hp-server-lib/entity"
	"sync"
	"time"
)

var (
	safeRule      = sync.Map{}      // 缓存: key=safeId(int), value=[0]=规则字符串, [1]=最后查询时间(time.Time)
	queryLock     = sync.Mutex{}    // 防并发查库的锁
	queryInterval = 5 * time.Minute // 5分钟查询一次
)

// GetRule 极简版：查缓存→限频率查库→返回结果
func GetRule(safeId int) string {
	// 1. 先查缓存：有则直接返回
	if val, ok := safeRule.Load(safeId); ok {
		// 缓存格式：[0]是规则字符串，[1]是最后查询时间（不用结构体，直接用切片）
		cacheData := val.([2]interface{})
		return cacheData[0].(string)
	}

	// 2. 缓存无，加锁防并发查库
	queryLock.Lock()
	defer queryLock.Unlock()

	// 双重检查：防止解锁前已有其他协程写入缓存
	if val, ok := safeRule.Load(safeId); ok {
		cacheData := val.([2]interface{})
		return cacheData[0].(string)
	}

	// 3. 检查是否满足5分钟查询频率（首次查询允许）
	now := time.Now()
	var canQuery = true // 默认允许查库
	// 先看有没有历史查询记录（即使没查到规则，也会存时间）
	if val, ok := safeRule.Load(safeId); ok {
		cacheData := val.([2]interface{})
		lastTime := cacheData[1].(time.Time)
		// 不到5分钟，不查库
		if now.Sub(lastTime) < queryInterval {
			canQuery = false
		}
	}
	if !canQuery {
		return ""
	}
	rule := "" // 默认返回空
	user := &entity.UserSafeEntity{}
	// 查不到数据时，Error非空，规则保持空
	if err := db.DB.Where("id=?", safeId).First(user).Error; err == nil {
		rule = user.Rule
	}

	// 6. 存缓存：不管有没有查到，都记录“规则+最后查询时间”
	safeRule.Store(safeId, [2]interface{}{rule, now})

	return rule
}

type UserSafeService struct {
}

func (receiver *UserSafeService) AddData(userId int, custom entity.UserSafeEntity) error {
	wafConfig := coraza.NewWAFConfig().WithDirectives(custom.Rule)
	_, err := coraza.NewWAF(wafConfig)
	if err != nil {
		return err
	}
	custom.UserId = userId
	db.DB.Save(&custom)
	safeRule.Delete(custom.Id)
	return nil
}

func (receiver *UserSafeService) ListData(userId int, page int, pageSize int) *bean.ResPage {
	var results []*entity.UserSafeEntity
	var total int64
	// 计算总记录数并执行分页查询
	if userId < 0 {
		db.DB.Model(&entity.UserSafeEntity{}).Order("id desc").Count(&total).Offset((page - 1) * pageSize).Limit(pageSize).Find(&results)
	} else {
		db.DB.Model(&entity.UserSafeEntity{}).Where("user_id = ?", userId).Order("id desc").Count(&total).Offset((page - 1) * pageSize).Limit(pageSize).Find(&results)
	}

	if userId < 0 {
		var userIds []int
		for _, item := range results {
			userIds = append(userIds, item.UserId)
		}
		var users []*entity.UserCustomEntity
		if err := db.DB.Model(&entity.UserCustomEntity{}).Where("id IN ?", userIds).Find(&users).Error; err == nil {
			// 将查询结果转换成 map[int]User
			userMap := make(map[int]*entity.UserCustomEntity)
			for _, user := range users {
				userMap[*user.Id] = user
			}
			for _, item := range results {
				customEntity := userMap[item.UserId]
				if customEntity != nil {
					item.Username = customEntity.Username
					item.UserDesc = customEntity.Desc
				}
			}
		}
	}

	return bean.PageOk(total, results)
}

func (receiver *UserSafeService) RemoveData(id int) {
	userQuery := &entity.UserSafeEntity{}
	db.DB.Where("id = ? ", id).First(userQuery)
	if userQuery != nil {
		db.DB.Delete(&entity.UserSafeEntity{Id: &id})
	}
	safeRule.Delete(id)

}

func (receiver *UserSafeService) SafeListByKey(userId int, keyword string) *bean.ResData {
	var results []*entity.UserSafeEntity
	if userId < 0 {
		tx := db.DB.Model(&entity.UserSafeEntity{})
		if len(keyword) > 0 {
			tx.Where("safe_name like ?", "%"+keyword+"%")
		}
		tx.Order("id desc").Find(&results)
	} else {
		model := db.DB.Model(&entity.UserSafeEntity{})
		if len(keyword) > 0 {
			model.Where("safe_name like ? and user_id = ? ", "%"+keyword+"%", userId)
		} else {
			model.Where("user_id = ?", userId)
		}
		model.Order("id desc").Find(&results)
	}
	for _, result := range results {
		result.Rule = ""
	}
	return bean.ResOk(results)
}
