// Package gomailer 提供了一个简单易用的邮件发送客户端
// 支持 SMTP 和 Sendmail 两种发送方式，可以发送 HTML/文本邮件，支持附件和内联附件
package gomailer

import (
	"bytes"
	"io"
	"net/mail"

	"github.com/gabriel-vasile/mimetype"
)

// Message 定义了一个通用的邮件消息结构体
// 包含邮件发送所需的全部信息：发件人、收件人、主题、正文、附件等
type Message struct {
	// From 发件人地址和姓名
	From mail.Address `json:"from"`

	// To 收件人列表（主要收件人）
	To []mail.Address `json:"to"`

	// Bcc 密送收件人列表（其他收件人看不到此列表）
	Bcc []mail.Address `json:"bcc"`

	// Cc 抄送收件人列表（其他收件人可以看到此列表）
	Cc []mail.Address `json:"cc"`

	// Subject 邮件主题
	Subject string `json:"subject"`

	// HTML HTML格式的邮件正文
	HTML string `json:"html"`

	// Text 纯文本格式的邮件正文（如果不提供，会自动从HTML转换）
	Text string `json:"text"`

	// Headers 自定义邮件头部信息
	Headers map[string]string `json:"headers"`

	// Attachments 普通附件（文件名 -> 文件内容读取器）
	Attachments map[string]io.Reader `json:"attachments"`

	// InlineAttachments 内联附件（通常用于在HTML中嵌入图片）
	InlineAttachments map[string]io.Reader `json:"inlineAttachments"`
}

// Mailer 定义了邮件客户端的基础接口
// 任何实现了 Send 方法的类型都可以作为邮件发送客户端
type Mailer interface {
	// Send 发送一封邮件
	// 参数:
	//   - message: 要发送的邮件消息
	// 返回:
	//   - error: 发送失败时返回错误信息，成功返回 nil
	Send(message *Message) error
}

// SendInterceptor 是一个可选接口，用于注册邮件发送钩子
// 实现此接口可以在邮件发送前后执行自定义逻辑
type SendInterceptor interface {
	// OnSend 返回发送钩子，可以在发送前后添加自定义处理
	OnSend() *Hook[*SendEvent]
}

// SendEvent 发送事件，包含发送过程中的邮件消息
type SendEvent struct {
	Event
	// Message 正在发送的邮件消息
	Message *Message
}

// addressesToStrings 将邮件地址列表转换为字符串列表
// 
// 参数:
//   - addresses: 邮件地址列表
//   - withName: 是否包含姓名（true 返回 "姓名 <email>"，false 只返回 "email"）
// 返回:
//   - []string: 转换后的字符串列表
func addressesToStrings(addresses []mail.Address, withName bool) []string {
	result := make([]string, len(addresses))

	for i, addr := range addresses {
		if withName && addr.Name != "" {
			result[i] = addr.String()
		} else {
			// 只保留邮箱部分，避免被包裹在尖括号中
			result[i] = addr.Address
		}
	}

	return result
}

// detectReaderMimeType 读取 Reader 的前几个字节来检测其 MIME 类型
// 这对于正确设置附件的内容类型很重要
//
// 参数:
//   - r: 要检测的数据流
// 返回:
//   - io.Reader: 一个新的组合 Reader（包含已读取的部分 + 原始 Reader 的剩余部分）
//   - string: 检测到的 MIME 类型（如 "image/png"、"application/pdf" 等）
//   - error: 检测失败时返回错误
func detectReaderMimeType(r io.Reader) (io.Reader, string, error) {
	readCopy := new(bytes.Buffer)

	// 使用 TeeReader 在读取的同时复制数据
	mime, err := mimetype.DetectReader(io.TeeReader(r, readCopy))
	if err != nil {
		return nil, "", err
	}

	// 返回一个组合 Reader：已读取的数据 + 原始 Reader 的剩余数据
	return io.MultiReader(readCopy, r), mime.String(), nil
}

