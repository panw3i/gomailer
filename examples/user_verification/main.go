package main

import (
	"fmt"
	"log"
	"net/mail"

	"github.com/yourusername/gomailer"
)

// 示例：用户注册验证邮件
// 这是一个常见的实际应用场景

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPTls      bool
	FromName     string
	FromAddress  string
}

// UserVerificationService 用户验证邮件服务
type UserVerificationService struct {
	config EmailConfig
	client gomailer.Mailer
}

// NewUserVerificationService 创建用户验证服务
func NewUserVerificationService(config EmailConfig) *UserVerificationService {
	client := &gomailer.SMTPClient{
		Host:     config.SMTPHost,
		Port:     config.SMTPPort,
		Username: config.SMTPUsername,
		Password: config.SMTPPassword,
		TLS:      config.SMTPTls,
	}

	return &UserVerificationService{
		config: config,
		client: client,
	}
}

// SendVerificationEmail 发送验证邮件
func (s *UserVerificationService) SendVerificationEmail(
	userEmail string,
	userName string,
	token string,
) error {
	// 构建验证链接
	verifyURL := fmt.Sprintf("https://example.com/verify?token=%s", token)

	// 构建 HTML 邮件内容
	htmlContent := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<style>
				body {
					font-family: Arial, sans-serif;
					line-height: 1.6;
					color: #333;
				}
				.container {
					max-width: 600px;
					margin: 0 auto;
					padding: 20px;
				}
				.header {
					background-color: #4CAF50;
					color: white;
					padding: 20px;
					text-align: center;
					border-radius: 5px 5px 0 0;
				}
				.content {
					background-color: #f9f9f9;
					padding: 30px;
					border-radius: 0 0 5px 5px;
				}
				.button {
					display: inline-block;
					padding: 12px 30px;
					background-color: #4CAF50;
					color: white !important;
					text-decoration: none;
					border-radius: 5px;
					margin: 20px 0;
				}
				.footer {
					text-align: center;
					padding: 20px;
					color: #666;
					font-size: 12px;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>欢迎注册！</h1>
				</div>
				<div class="content">
					<p>你好 %s，</p>
					<p>感谢你注册我们的服务！为了完成注册，请点击下面的按钮验证你的邮箱地址：</p>
					<p style="text-align: center;">
						<a href="%s" class="button">验证邮箱</a>
					</p>
					<p>或者复制以下链接到浏览器中打开：</p>
					<p style="word-break: break-all; background-color: #fff; padding: 10px; border: 1px solid #ddd;">
						%s
					</p>
					<p><strong>注意：</strong>此验证链接将在 24 小时后过期。</p>
					<p>如果你没有注册账号，请忽略此邮件。</p>
				</div>
				<div class="footer">
					<p>这是一封系统自动发送的邮件，请勿直接回复。</p>
					<p>&copy; 2025 Example Company. All rights reserved.</p>
				</div>
			</div>
		</body>
		</html>
	`, userName, verifyURL, verifyURL)

	// 构建邮件消息
	message := &gomailer.Message{
		From: mail.Address{
			Name:    s.config.FromName,
			Address: s.config.FromAddress,
		},
		To: []mail.Address{
			{
				Name:    userName,
				Address: userEmail,
			},
		},
		Subject: "请验证您的邮箱地址",
		HTML:    htmlContent,
	}

	// 发送邮件
	return s.client.Send(message)
}

// SendPasswordResetEmail 发送密码重置邮件
func (s *UserVerificationService) SendPasswordResetEmail(
	userEmail string,
	userName string,
	resetToken string,
) error {
	// 构建重置链接
	resetURL := fmt.Sprintf("https://example.com/reset-password?token=%s", resetToken)

	// 构建 HTML 邮件内容
	htmlContent := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<style>
				body {
					font-family: Arial, sans-serif;
					line-height: 1.6;
					color: #333;
				}
				.container {
					max-width: 600px;
					margin: 0 auto;
					padding: 20px;
				}
				.header {
					background-color: #FF9800;
					color: white;
					padding: 20px;
					text-align: center;
					border-radius: 5px 5px 0 0;
				}
				.content {
					background-color: #f9f9f9;
					padding: 30px;
					border-radius: 0 0 5px 5px;
				}
				.button {
					display: inline-block;
					padding: 12px 30px;
					background-color: #FF9800;
					color: white !important;
					text-decoration: none;
					border-radius: 5px;
					margin: 20px 0;
				}
				.warning {
					background-color: #fff3cd;
					border-left: 4px solid #ffc107;
					padding: 12px;
					margin: 15px 0;
				}
				.footer {
					text-align: center;
					padding: 20px;
					color: #666;
					font-size: 12px;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>密码重置请求</h1>
				</div>
				<div class="content">
					<p>你好 %s，</p>
					<p>我们收到了重置你账号密码的请求。</p>
					<p>如果这是你本人的操作，请点击下面的按钮重置密码：</p>
					<p style="text-align: center;">
						<a href="%s" class="button">重置密码</a>
					</p>
					<p>或者复制以下链接到浏览器中打开：</p>
					<p style="word-break: break-all; background-color: #fff; padding: 10px; border: 1px solid #ddd;">
						%s
					</p>
					<div class="warning">
						<strong>⚠️ 安全提示：</strong>
						<ul>
							<li>此重置链接将在 1 小时后过期</li>
							<li>为了您的账号安全，请不要将此链接分享给他人</li>
							<li>如果您没有请求重置密码，请忽略此邮件</li>
						</ul>
					</div>
				</div>
				<div class="footer">
					<p>这是一封系统自动发送的邮件，请勿直接回复。</p>
					<p>&copy; 2025 Example Company. All rights reserved.</p>
				</div>
			</div>
		</body>
		</html>
	`, userName, resetURL, resetURL)

	// 构建邮件消息
	message := &gomailer.Message{
		From: mail.Address{
			Name:    s.config.FromName,
			Address: s.config.FromAddress,
		},
		To: []mail.Address{
			{
				Name:    userName,
				Address: userEmail,
			},
		},
		Subject: "重置您的密码",
		HTML:    htmlContent,
	}

	// 发送邮件
	return s.client.Send(message)
}

func main() {
	// 配置邮件服务
	config := EmailConfig{
		SMTPHost:     "smtp.gmail.com",
		SMTPPort:     587,
		SMTPUsername: "your-email@gmail.com",
		SMTPPassword: "your-app-password",
		SMTPTls:      true,
		FromName:     "Example 团队",
		FromAddress:  "noreply@example.com",
	}

	// 创建服务实例
	service := NewUserVerificationService(config)

	// 示例1：发送验证邮件
	log.Println("发送验证邮件...")
	err := service.SendVerificationEmail(
		"user@example.com",
		"张三",
		"example_verification_token_123456",
	)
	if err != nil {
		log.Printf("验证邮件发送失败: %v\n", err)
	} else {
		log.Println("✅ 验证邮件发送成功！")
	}

	// 示例2：发送密码重置邮件
	log.Println("\n发送密码重置邮件...")
	err = service.SendPasswordResetEmail(
		"user@example.com",
		"张三",
		"example_reset_token_789012",
	)
	if err != nil {
		log.Printf("密码重置邮件发送失败: %v\n", err)
	} else {
		log.Println("✅ 密码重置邮件发送成功！")
	}
}
