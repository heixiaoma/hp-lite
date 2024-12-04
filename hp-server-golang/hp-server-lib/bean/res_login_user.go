package bean

import (
	"hp-server-lib/util"
	"strconv"
	"time"
)

type ResLoginUser struct {
	Token   string `json:"token"`
	ExpTime int64  `json:"expTime"`
	Email   string `json:"email"`
	Role    Role   `json:"role"`
}

func NewAdminUser(user ReqLogin) *ResLoginUser {
	currentTime := time.Now()
	// 加上 3 天
	threeDaysLater := currentTime.Add(3 * 24 * time.Hour)
	// 获取三天后的 Unix 时间戳
	threeDaysLaterUnix := threeDaysLater.UnixMilli()
	token, _ := util.GenerateToken(strconv.Itoa(-1), "ADMIN")
	return &ResLoginUser{
		ExpTime: threeDaysLaterUnix,
		Email:   user.Email,
		Role:    ADMIN,
		Token:   token,
	}
}

func NewClientUser(userId int, email string) *ResLoginUser {
	currentTime := time.Now()
	// 加上 3 天
	threeDaysLater := currentTime.Add(3 * 24 * time.Hour)
	// 获取三天后的 Unix 时间戳
	threeDaysLaterUnix := threeDaysLater.UnixMilli()
	token, _ := util.GenerateToken(strconv.Itoa(userId), "CLIENT")
	return &ResLoginUser{
		ExpTime: threeDaysLaterUnix,
		Email:   email,
		Role:    CLIENT,
		Token:   token,
	}
}
