package main

import (
	"log"
	"net/mail"
	"time"

	"github.com/yourusername/gomailer"
)

// 示例：使用钩子函数记录邮件发送过程
func main() {
	// 创建 SMTP 客户端
	client := &gomailer.SMTPClient{
		Host:     "smtp.gmail.com",
		Port:     587,
		Username: "your-email@gmail.com",
		Password: "your-app-password",
		TLS:      true,
	}

	// 添加第一个钩子：记录发送开始时间
	client.OnSend().BindFunc(func(e *gomailer.SendEvent) error {
		log.Printf("📤 准备发送邮件")
		log.Printf("   主题: %s", e.Message.Subject)
		log.Printf("   发件人: %s", e.Message.From.Address)
		if len(e.Message.To) > 0 {
			log.Printf("   收件人: %s", e.Message.To[0].Address)
		}
		
		// 继续执行发送流程
		return e.Next()
	})

	// 添加第二个钩子：统计发送耗时
	client.OnSend().BindFunc(func(e *gomailer.SendEvent) error {
		start := time.Now()
		
		log.Println("⏱️  开始计时...")
		
		// 执行实际的发送操作
		err := e.Next()
		
		// 计算耗时
		duration := time.Since(start)
		
		if err != nil {
			log.Printf("❌ 发送失败 (耗时: %v): %v", duration, err)
		} else {
			log.Printf("✅ 发送成功 (耗时: %v)", duration)
		}
		
		return err
	})

	// 添加第三个钩子：记录到数据库（示例）
	client.OnSend().BindFunc(func(e *gomailer.SendEvent) error {
		// 先执行发送
		err := e.Next()
		
		// 发送完成后记录日志
		if err != nil {
			log.Println("📝 记录到日志系统: 发送失败")
			// 这里可以写入数据库或日志文件
			// saveToDatabase("failed", e.Message.Subject, err.Error())
		} else {
			log.Println("📝 记录到日志系统: 发送成功")
			// saveToDatabase("success", e.Message.Subject, "")
		}
		
		return err
	})

	// 构建邮件消息
	message := &gomailer.Message{
		From: mail.Address{
			Name:    "系统通知",
			Address: "noreply@example.com",
		},
		To: []mail.Address{
			{
				Name:    "用户",
				Address: "user@example.com",
			},
		},
		Subject: "这是一封带钩子的测试邮件",
		HTML: `
			<h2>钩子函数演示</h2>
			<p>这封邮件演示了如何使用钩子函数：</p>
			<ul>
				<li>在发送前记录邮件信息</li>
				<li>统计发送耗时</li>
				<li>发送后记录日志</li>
			</ul>
			<p>查看控制台输出以查看钩子函数的执行过程。</p>
		`,
	}

	// 发送邮件（钩子会自动执行）
	log.Println("================== 开始发送 ==================")
	if err := client.Send(message); err != nil {
		log.Fatal("发送过程出错:", err)
	}
	log.Println("================== 发送完成 ==================")
}

