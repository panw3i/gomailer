package main

import (
	"log"
	"net/mail"

	"github.com/yourusername/gomailer"
)

// 基础示例：发送简单的 HTML 邮件
func main() {
	// 创建 SMTP 客户端
	// 注意：这里使用的是示例配置，实际使用时请替换为你的 SMTP 服务器信息
	client := &gomailer.SMTPClient{
		Host:     "smtp.gmail.com",        // SMTP 服务器地址
		Port:     587,                      // SMTP 端口（587 用于 STARTTLS）
		Username: "your-email@gmail.com",   // 你的邮箱地址
		Password: "your-app-password",      // 你的应用专用密码（不是账号密码）
		TLS:      true,                     // 使用 TLS 加密
	}

	// 构建邮件消息
	message := &gomailer.Message{
		From: mail.Address{
			Name:    "发件人名称",              // 发件人显示的名称
			Address: "sender@example.com",   // 发件人邮箱地址
		},
		To: []mail.Address{
			{
				Name:    "收件人名称",            // 收件人显示的名称（可选）
				Address: "recipient@example.com", // 收件人邮箱地址
			},
		},
		Subject: "这是一封测试邮件",              // 邮件主题
		HTML:    "<h1>你好！</h1><p>这是一封使用 GoMailer 发送的测试邮件。</p><p>邮件发送时间：2025年</p>",
	}

	// 发送邮件
	log.Println("正在发送邮件...")
	if err := client.Send(message); err != nil {
		log.Fatal("发送失败:", err)
	}

	log.Println("✅ 邮件发送成功！")
}

