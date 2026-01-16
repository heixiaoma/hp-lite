package service

import (
	"hp-server-lib/bean"
	"hp-server-lib/db"
	"hp-server-lib/entity"
	"time"
)

type UserCustomService struct {
	passwordService *PasswordService
}

func (receiver *UserCustomService) AddData(custom entity.UserCustomEntity) error {
	// 初始化passwordService
	if receiver.passwordService == nil {
		receiver.passwordService = &PasswordService{}
	}

	// 如果是新增用户
	if custom.Id == nil {
		custom.CreateTime = time.Now()
		// 新增用户必须有密码
		if custom.Password == "" {
			return nil // 密码为空，不保存
		}
		// 加密密码
		hashedPassword, err := receiver.passwordService.HashPassword(custom.Password)
		if err != nil {
			return err
		}
		custom.Password = hashedPassword
	} else {
		// 如果是编辑用户
		// 如果密码不为空，则更新密码
		if custom.Password != "" {
			hashedPassword, err := receiver.passwordService.HashPassword(custom.Password)
			if err != nil {
				return err
			}
			custom.Password = hashedPassword
		} else {
			// 如果密码为空，保持原密码不变
			var existingUser entity.UserCustomEntity
			db.DB.Where("id = ?", custom.Id).First(&existingUser)
			custom.Password = existingUser.Password
		}
	}

	db.DB.Save(&custom)
	return nil
}

func (receiver *UserCustomService) ListData(page int, pageSize int) *bean.ResPage {
	var results []entity.UserCustomEntity
	var total int64
	// 计算总记录数并执行分页查询
	db.DB.Model(&entity.UserCustomEntity{}).Order("id desc").Count(&total).Offset((page - 1) * pageSize).Limit(pageSize).Find(&results)
	return bean.PageOk(total, results)
}

func (receiver *UserCustomService) RemoveData(id int) {
	db.DB.Delete(&entity.UserCustomEntity{Id: &id})
}
