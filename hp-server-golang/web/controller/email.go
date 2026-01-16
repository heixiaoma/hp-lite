package controller

import (
	"encoding/json"
	"hp-server-lib/bean"
	"hp-server-lib/config"
	"hp-server-lib/db"
	"hp-server-lib/entity"
	"hp-server-lib/service"
	"net/http"
)

type EmailController struct {
	emailCodeService *service.EmailCodeService
	emailService     *service.EmailService
	passwordService  *service.PasswordService
}

func NewEmailController() *EmailController {
	return &EmailController{
		emailCodeService: &service.EmailCodeService{},
		emailService:     &service.EmailService{},
		passwordService:  &service.PasswordService{},
	}
}

// 发送验证码
func (ec *EmailController) SendCode(w http.ResponseWriter, r *http.Request) {
	var req bean.ReqSendCode
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(bean.ResError("请求参数错误"))
		return
	}

	if req.Email == "" {
		json.NewEncoder(w).Encode(bean.ResError("邮箱不能为空"))
		return
	}

	if req.Type == "" {
		json.NewEncoder(w).Encode(bean.ResError("验证码类型不能为空"))
		return
	}

	// 如果是重置密码，检查邮箱是否存在
	if req.Type == "reset_password" {
		var user entity.UserCustomEntity
		if err := db.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
			json.NewEncoder(w).Encode(bean.ResError("该邮箱未注册"))
			return
		}
	}

	_, err := ec.emailCodeService.GenerateCode(req.Email, req.Type, nil)
	if err != nil {
		json.NewEncoder(w).Encode(bean.ResError("发送验证码失败: " + err.Error()))
		return
	}

	json.NewEncoder(w).Encode(bean.ResOk(map[string]string{
		"message": "验证码已发送到您的邮箱，请在30分钟内使用",
	}))
}

// 验证邮箱验证码
func (ec *EmailController) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req bean.ReqVerifyCode
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(bean.ResError("请求参数错误"))
		return
	}

	if req.Email == "" || req.Code == "" {
		json.NewEncoder(w).Encode(bean.ResError("邮箱和验证码不能为空"))
		return
	}

	codeType := "verify_email"
	if req.Type != "" {
		codeType = req.Type
	}

	var valid bool
	var err error
	if codeType == "reset_password" {
		valid, err = ec.emailCodeService.CheckCode(req.Email, req.Code, codeType)
	} else {
		valid, err = ec.emailCodeService.VerifyCode(req.Email, req.Code, codeType)
	}
	if err != nil {
		json.NewEncoder(w).Encode(bean.ResError("验证失败: " + err.Error()))
		return
	}

	if !valid {
		json.NewEncoder(w).Encode(bean.ResError("验证码无效或已过期"))
		return
	}

	json.NewEncoder(w).Encode(bean.ResOk(map[string]string{
		"message": "邮箱验证成功",
	}))
}

// 密码重置 - 验证验证码并重置密码
func (ec *EmailController) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req bean.ReqResetPassword
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(bean.ResError("请求参数错误"))
		return
	}

	if req.Email == "" || req.Code == "" || req.Password == "" {
		json.NewEncoder(w).Encode(bean.ResError("邮箱、验证码和密码不能为空"))
		return
	}

	if len(req.Password) < 6 {
		json.NewEncoder(w).Encode(bean.ResError("密码长度不能少于6位"))
		return
	}

	// 验证验证码
	valid, err := ec.emailCodeService.VerifyCode(req.Email, req.Code, "reset_password")
	if err != nil {
		json.NewEncoder(w).Encode(bean.ResError("验证失败: " + err.Error()))
		return
	}

	if !valid {
		json.NewEncoder(w).Encode(bean.ResError("验证码无效或已过期"))
		return
	}

	// 加密新密码
	hashedPassword, err := ec.passwordService.HashPassword(req.Password)
	if err != nil {
		json.NewEncoder(w).Encode(bean.ResError("密码加密失败"))
		return
	}

	// 更新用户密码
	if err := db.DB.Model(&entity.UserCustomEntity{}).Where("email = ?", req.Email).Update("password", hashedPassword).Error; err != nil {
		json.NewEncoder(w).Encode(bean.ResError("重置密码失败"))
		return
	}

	json.NewEncoder(w).Encode(bean.ResOk(map[string]string{
		"message": "密码重置成功，请用新密码登录",
	}))
}

// 设置用户邮箱（个人设置）
func (ec *EmailController) SetUserEmail(w http.ResponseWriter, r *http.Request) {
	var req bean.ReqSetEmail
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(bean.ResError("请求参数错误"))
		return
	}

	// 从请求头中获取用户ID（假设已验证）
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		json.NewEncoder(w).Encode(bean.ResError("未登录"))
		return
	}

	// 管理员不支持修改邮箱
	if userIDStr == "-1" {
		json.NewEncoder(w).Encode(bean.ResError("管理员账户不支持修改邮箱"))
		return
	}

	if req.Email == "" || req.Code == "" {
		json.NewEncoder(w).Encode(bean.ResError("邮箱和验证码不能为空"))
		return
	}

	// 验证验证码
	valid, err := ec.emailCodeService.VerifyCode(req.Email, req.Code, "verify_email")
	if err != nil {
		json.NewEncoder(w).Encode(bean.ResError("验证失败: " + err.Error()))
		return
	}

	if !valid {
		json.NewEncoder(w).Encode(bean.ResError("验证码无效或已过期"))
		return
	}

	// 检查邮箱是否已被使用
	var count int64
	db.DB.Model(&entity.UserCustomEntity{}).Where("email = ? AND id != ?", req.Email, userIDStr).Count(&count)
	if count > 0 {
		json.NewEncoder(w).Encode(bean.ResError("该邮箱已被其他用户使用"))
		return
	}

	// 更新用户邮箱
	if err := db.DB.Model(&entity.UserCustomEntity{}).Where("id = ?", userIDStr).Update("email", req.Email).Error; err != nil {
		json.NewEncoder(w).Encode(bean.ResError("设置邮箱失败: " + err.Error()))
		return
	}

	json.NewEncoder(w).Encode(bean.ResOk(map[string]string{
		"message": "邮箱设置成功",
	}))
}

// 获取用户邮箱（个人信息）
func (ec *EmailController) GetUserEmail(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		json.NewEncoder(w).Encode(bean.ResError("未登录"))
		return
	}

	// 如果是管理员用户（ID为-1），返回配置文件中的邮箱
	if userIDStr == "-1" {
		json.NewEncoder(w).Encode(bean.ResOk(map[string]string{
			"email": config.ConfigData.Admin.Username,
		}))
		return
	}

	var user entity.UserCustomEntity
	if err := db.DB.Where("id = ?", userIDStr).First(&user).Error; err != nil {
		json.NewEncoder(w).Encode(bean.ResError("获取用户信息失败: " + err.Error()))
		return
	}

	json.NewEncoder(w).Encode(bean.ResOk(map[string]string{
		"email": user.Email,
	}))
}

// 修改密码（个人设置）
func (ec *EmailController) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req bean.ReqChangePassword
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(bean.ResError("请求参数错误"))
		return
	}

	// 从请求头中获取用户ID
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		json.NewEncoder(w).Encode(bean.ResError("未登录"))
		return
	}

	// 管理员修改密码需要修改配置文件，这里不支持
	if userIDStr == "-1" {
		json.NewEncoder(w).Encode(bean.ResError("管理员密码请在配置文件中修改"))
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		json.NewEncoder(w).Encode(bean.ResError("旧密码和新密码不能为空"))
		return
	}

	if len(req.NewPassword) < 6 {
		json.NewEncoder(w).Encode(bean.ResError("新密码长度不能少于6位"))
		return
	}

	// 获取用户信息
	var user entity.UserCustomEntity
	if err := db.DB.Where("id = ?", userIDStr).First(&user).Error; err != nil {
		json.NewEncoder(w).Encode(bean.ResError("获取用户信息失败: " + err.Error()))
		return
	}

	// 验证旧密码
	if !ec.passwordService.VerifyPassword(user.Password, req.OldPassword) {
		json.NewEncoder(w).Encode(bean.ResError("当前密码错误"))
		return
	}

	// 加密新密码
	hashedPassword, err := ec.passwordService.HashPassword(req.NewPassword)
	if err != nil {
		json.NewEncoder(w).Encode(bean.ResError("密码加密失败: " + err.Error()))
		return
	}

	// 更新密码
	if err := db.DB.Model(&entity.UserCustomEntity{}).Where("id = ?", userIDStr).Update("password", hashedPassword).Error; err != nil {
		json.NewEncoder(w).Encode(bean.ResError("修改密码失败: " + err.Error()))
		return
	}

	json.NewEncoder(w).Encode(bean.ResOk(map[string]string{
		"message": "密码修改成功",
	}))
}

