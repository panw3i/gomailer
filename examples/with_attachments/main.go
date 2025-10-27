package main

import (
	"bytes"
	"io"
	"log"
	"net/mail"

	"github.com/yourusername/gomailer"
)

// 示例：发送带附件的邮件
func main() {
	// 创建 SMTP 客户端
	client := &gomailer.SMTPClient{
		Host:     "smtp.gmail.com",
		Port:     587,
		Username: "your-email@gmail.com",
		Password: "your-app-password",
		TLS:      true,
	}

	// 创建一个模拟的 PDF 文件内容
	// 在实际使用中，你可以使用 os.Open() 打开真实的文件
	pdfContent := []byte("%PDF-1.4\n这是一个模拟的 PDF 文件内容")
	pdfReader := bytes.NewReader(pdfContent)

	// 创建一个模拟的文本文件
	txtContent := []byte("这是一个附件文本文件的内容\n包含一些测试数据")
	txtReader := bytes.NewReader(txtContent)

	// 构建邮件消息
	message := &gomailer.Message{
		From: mail.Address{
			Name:    "发件人",
			Address: "sender@example.com",
		},
		To: []mail.Address{
			{Address: "recipient@example.com"},
		},
		Subject: "带附件的邮件示例",
		HTML: `
			<h2>附件邮件测试</h2>
			<p>这封邮件包含两个附件：</p>
			<ol>
				<li>一个 PDF 文档</li>
				<li>一个文本文件</li>
			</ol>
			<p>请查收附件。</p>
		`,
		// 添加附件
		Attachments: map[string]io.Reader{
			"document.pdf": pdfReader, // 文件名 -> 文件内容
			"readme.txt":   txtReader,
		},
	}

	// 发送邮件
	log.Println("正在发送带附件的邮件...")
	if err := client.Send(message); err != nil {
		log.Fatal("发送失败:", err)
	}

	log.Println("✅ 带附件的邮件发送成功！")
}
