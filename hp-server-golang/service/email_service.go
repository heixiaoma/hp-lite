package service

import (
	"fmt"
	"hp-server-lib/config"
	"math/rand"
	"net/smtp"
	"strconv"
	"time"
)

type EmailService struct {
}

// 发送验证码邮件
func (e *EmailService) SendVerificationCode(toEmail string) (string, error) {
	code := e.generateCode()
	subject := "HP-Lite 邮箱验证码"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; }
        .container { max-width: 600px; margin: 0 auto; }
        .header { background: linear-gradient(135deg, #4b6ff6 0%%, #1890ff 100%%); color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f5f5f5; padding: 20px; }
        .code { font-size: 32px; font-weight: bold; color: #1890ff; text-align: center; padding: 20px; background: white; margin: 20px 0; border-radius: 5px; }
        .footer { background: white; padding: 20px; text-align: center; color: #666; border-radius: 0 0 5px 5px; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>HP-Lite 邮箱验证</h1>
        </div>
        <div class="content">
            <p>尊敬的用户，您好！</p>
            <p>您正在进行邮箱验证，请使用下方验证码完成验证，有效期为30分钟。</p>
            <div class="code">%s</div>
            <p style="color: #ff0000;">如果您没有请求此验证，请忽略此邮件。</p>
        </div>
        <div class="footer">
            <p>© 2024 HP-Lite. 内网穿透解决方案</p>
            <p>这是一封自动发送的邮件，请勿直接回复</p>
        </div>
    </div>
</body>
</html>
`, code)
	err := e.sendEmail(toEmail, subject, body)
	if err != nil {
		return "", err
	}
	return code, nil
}

// 发送密码重置邮件
func (e *EmailService) SendPasswordReset(toEmail string, resetCode string) error {
	subject := "HP-Lite 密码重置确认"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; }
        .container { max-width: 600px; margin: 0 auto; }
        .header { background: linear-gradient(135deg, #4b6ff6 0%%, #1890ff 100%%); color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f5f5f5; padding: 20px; }
        .code { font-size: 32px; font-weight: bold; color: #1890ff; text-align: center; padding: 20px; background: white; margin: 20px 0; border-radius: 5px; }
        .footer { background: white; padding: 20px; text-align: center; color: #666; border-radius: 0 0 5px 5px; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>HP-Lite 密码重置</h1>
        </div>
        <div class="content">
            <p>尊敬的用户，您好！</p>
            <p>您正在重置账户密码，请使用下方验证码完成密码重置，有效期为30分钟。</p>
            <div class="code">%s</div>
            <p style="color: #ff0000;">如果您没有请求此操作，请忽略此邮件并确保您的账户安全。</p>
        </div>
        <div class="footer">
            <p>© 2024 HP-Lite. 内网穿透解决方案</p>
            <p>这是一封自动发送的邮件，请勿直接回复</p>
        </div>
    </div>
</body>
</html>
`, resetCode)
	return e.sendEmail(toEmail, subject, body)
}

// 内部方法：发送邮件
func (e *EmailService) sendEmail(to, subject, body string) error {
	cfg := config.ConfigData.Email

	from := cfg.From
	password := cfg.Password
	smtpHost := cfg.SmtpHost
	smtpPort := strconv.Itoa(cfg.SmtpPort)

	// 验证配置
	if from == "" || password == "" || smtpHost == "" {
		return fmt.Errorf("邮件配置不完整，请检查配置文件中的email配置项")
	}

	auth := smtp.PlainAuth("", from, password, smtpHost)

	msg := fmt.Sprintf("From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		cfg.FromName, from, to, subject, body)

	addr := smtpHost + ":" + smtpPort
	err := smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("SMTP发送失败 (%s:%s): %v", smtpHost, smtpPort, err)
	}
	return nil
}

// 生成6位随机验证码
func (e *EmailService) generateCode() string {
	rand.Seed(time.Now().UnixNano())
	code := rand.Intn(900000) + 100000
	return strconv.Itoa(code)
}
