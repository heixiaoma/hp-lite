package service

import (
	"hp-server-lib/bean"
	"hp-server-lib/config"
	"hp-server-lib/db"
	"hp-server-lib/entity"
	"strings"
)

type UserService struct {
}

func (receiver *UserService) Login(login bean.ReqLogin) *bean.ResLoginUser {
	if strings.Compare(config.AdminUser, login.Email) == 0 && strings.Compare(config.AdminPassword, login.Password) == 0 {
		return bean.NewAdminUser(login)
	} else {
		userQuery := entity.UserCustomEntity{
			Username: login.Email,
			Password: login.Password,
		}
		db.DB.Find(&userQuery)
		if userQuery.Id > 0 {
			return bean.NewClientUser(userQuery.Id, userQuery.Username)
		}
	}
	return nil
}
