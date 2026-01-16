package service

import (
	"hp-server-lib/bean"
	"hp-server-lib/config"
	"hp-server-lib/db"
	"hp-server-lib/entity"
	"strings"
)

type UserService struct {
	passwordService *PasswordService
}

func (receiver *UserService) Login(login bean.ReqLogin) *bean.ResLoginUser {
	if strings.Compare(config.ConfigData.Admin.Username, login.Email) == 0 && strings.Compare(config.ConfigData.Admin.Password, login.Password) == 0 {
		return bean.NewAdminUser(login)
	} else {
		userQuery := entity.UserCustomEntity{}
		db.DB.Where("username = ?", login.Email).First(&userQuery)
		if userQuery.Id != nil {
			// 初始化passwordService
			if receiver.passwordService == nil {
				receiver.passwordService = &PasswordService{}
			}
			// 验证密码
			if receiver.passwordService.VerifyPassword(userQuery.Password, login.Password) {
				return bean.NewClientUser(*userQuery.Id, userQuery.Username)
			}
		}
	}
	return nil
}
