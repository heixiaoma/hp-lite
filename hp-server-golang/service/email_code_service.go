package service

import (
	"hp-server-lib/db"
	"hp-server-lib/entity"
	"time"
)

type EmailCodeService struct {
}

// 生成并保存验证码
func (e *EmailCodeService) GenerateCode(email, codeType string, userId *int) (string, error) {
	emailService := &EmailService{}
	code, err := emailService.SendVerificationCode(email)
	if err != nil {
		return "", err
	}

	// 保存验证码到数据库
	expireTime := time.Now().Add(30 * time.Minute)
	emailCode := &entity.EmailCodeEntity{
		Email:      email,
		Code:       code,
		Type:       codeType,
		UserId:     userId,
		Used:       false,
		CreateTime: time.Now(),
		ExpireTime: expireTime,
	}

	if err := db.DB.Create(emailCode).Error; err != nil {
		return "", err
	}

	return code, nil
}

// 验证码是否有效
func (e *EmailCodeService) VerifyCode(email, code, codeType string) (bool, error) {
	var emailCode entity.EmailCodeEntity
	err := db.DB.Where("email = ? AND code = ? AND type = ? AND used = false", email, code, codeType).
		Where("expire_time > ?", time.Now()).
		First(&emailCode).Error

	if err != nil {
		return false, err
	}

	// 标记为已使用
	db.DB.Model(&emailCode).Update("used", true)
	return true, nil
}

// 清理过期的验证码
func (e *EmailCodeService) CleanupExpiredCodes() error {
	return db.DB.Where("expire_time < ?", time.Now()).Delete(&entity.EmailCodeEntity{}).Error
}
