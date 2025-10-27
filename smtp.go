package gomailer

import (
    "errors"
    "fmt"
    "net/smtp"
    "strings"

    "github.com/domodwyer/mailyak/v3"
)

// 确保 SMTPClient 实现了 Mailer 接口
var _ Mailer = (*SMTPClient)(nil)

const (
	// SMTPAuthPlain PLAIN 认证方法（默认）
	SMTPAuthPlain = "PLAIN"
	// SMTPAuthLogin LOGIN 认证方法（某些服务如 Outlook 需要）
	SMTPAuthLogin = "LOGIN"
)

// SMTPClient 定义了一个 SMTP 邮件客户端结构
// 实现了 Mailer 接口，可以通过 SMTP 协议发送邮件
type SMTPClient struct {
	// onSend 发送钩子，允许在发送前后执行自定义逻辑
	onSend *Hook[*SendEvent]

	// TLS 是否使用 TLS 加密连接
	TLS bool

	// Port SMTP 服务器端口（通常为 25、465 或 587）
	Port int

	// Host SMTP 服务器地址（如 "smtp.gmail.com"）
	Host string

	// Username SMTP 认证用户名
	Username string

	// Password SMTP 认证密码
	Password string

	// AuthMethod SMTP 认证方法
	// 如果未明确设置，默认使用 "PLAIN"
	// 可选值: SMTPAuthPlain, SMTPAuthLogin
	AuthMethod string

	// LocalName 用于 EHLO/HELO 交换的可选域名
	// 如果未明确设置，默认为 "localhost"
	// 某些 SMTP 服务器需要此设置，例如 Gmail SMTP-relay
	LocalName string
}

// OnSend 实现 SendInterceptor 接口
// 返回发送钩子，允许用户在邮件发送前后添加自定义处理逻辑
//
// 示例:
//   client := &SMTPClient{...}
//   client.OnSend().BindFunc(func(e *SendEvent) error {
//       fmt.Println("准备发送邮件:", e.Message.Subject)
//       return e.Next()
//   })
func (c *SMTPClient) OnSend() *Hook[*SendEvent] {
	if c.onSend == nil {
		c.onSend = &Hook[*SendEvent]{}
	}
	return c.onSend
}

// Send 实现 Mailer 接口
// 通过 SMTP 协议发送邮件
//
// 参数:
//   - m: 要发送的邮件消息
// 返回:
//   - error: 发送失败时返回错误，成功返回 nil
func (c *SMTPClient) Send(m *Message) error {
	if c.onSend != nil {
		return c.onSend.Trigger(&SendEvent{Message: m}, func(e *SendEvent) error {
			return c.send(e.Message)
		})
	}

	return c.send(m)
}

// send 内部发送方法，执行实际的 SMTP 发送操作
func (c *SMTPClient) send(m *Message) error {
    // 基础输入校验
    if m == nil {
        return errors.New("message is nil")
    }
    if m.From.Address == "" {
        return errors.New("from address is required")
    }
    if len(m.To) == 0 && len(m.Cc) == 0 && len(m.Bcc) == 0 {
        return errors.New("at least one recipient (To/Cc/Bcc) is required")
    }

    // 配置 SMTP 认证
    var smtpAuth smtp.Auth
    if c.Username != "" || c.Password != "" {
        if c.Username == "" || c.Password == "" {
            return errors.New("both username and password are required when using SMTP auth")
        }
        switch c.AuthMethod {
        case SMTPAuthLogin:
            // 使用 LOGIN 认证（某些服务如 Outlook 需要）
            smtpAuth = &smtpLoginAuth{c.Username, c.Password}
        default:
            // 默认使用 PLAIN 认证
            smtpAuth = smtp.PlainAuth("", c.Username, c.Password, c.Host)
        }
    }

    // 创建 MailYak 实例
    var yak *mailyak.MailYak
    if c.TLS {
        // 465 端口通常为隐式 TLS，其它端口（如 587）使用 STARTTLS
        if c.Port == 465 {
            var tlsErr error
            yak, tlsErr = mailyak.NewWithTLS(fmt.Sprintf("%s:%d", c.Host, c.Port), smtpAuth, nil)
            if tlsErr != nil {
                return tlsErr
            }
        } else {
            // 587/25 等端口：使用常规连接，由 MailYak 进行 STARTTLS（若可用）
            yak = mailyak.New(fmt.Sprintf("%s:%d", c.Host, c.Port), smtpAuth)
        }
    } else {
        // 明确关闭 TLS：使用常规连接
        yak = mailyak.New(fmt.Sprintf("%s:%d", c.Host, c.Port), smtpAuth)
    }

	// 设置本地主机名（如果指定）
	if c.LocalName != "" {
		yak.LocalName(c.LocalName)
	}

	// 设置发件人信息
	if m.From.Name != "" {
		yak.FromName(m.From.Name)
	}
	yak.From(m.From.Address)
	yak.Subject(m.Subject)
	yak.HTML().Set(m.HTML)

	// 设置纯文本内容
	if m.Text == "" {
		// 尝试从 HTML 自动生成纯文本版本
		if plain, err := html2Text(m.HTML); err == nil {
			yak.Plain().Set(plain)
		}
	} else {
		yak.Plain().Set(m.Text)
	}

	// 设置收件人
	if len(m.To) > 0 {
		yak.To(addressesToStrings(m.To, true)...)
	}

	// 设置密送收件人
	if len(m.Bcc) > 0 {
		yak.Bcc(addressesToStrings(m.Bcc, true)...)
	}

	// 设置抄送收件人
	if len(m.Cc) > 0 {
		yak.Cc(addressesToStrings(m.Cc, true)...)
	}

	// 添加普通附件
	for name, data := range m.Attachments {
		r, mime, err := detectReaderMimeType(data)
		if err != nil {
			return err
		}
		yak.AttachWithMimeType(name, r, mime)
	}

	// 添加内联附件（用于 HTML 中的嵌入图片等）
	for name, data := range m.InlineAttachments {
		r, mime, err := detectReaderMimeType(data)
		if err != nil {
			return err
		}
		yak.AttachInlineWithMimeType(name, r, mime)
	}

	// 添加自定义邮件头
	var hasMessageId bool
	for k, v := range m.Headers {
		if strings.EqualFold(k, "Message-ID") {
			hasMessageId = true
		}
		yak.AddHeader(k, v)
	}

	// 如果没有 Message-ID，添加一个默认的
	if !hasMessageId {
		fromParts := strings.Split(m.From.Address, "@")
		if len(fromParts) == 2 {
			yak.AddHeader("Message-ID", fmt.Sprintf("<%s@%s>",
				pseudorandomString(15),
				fromParts[1],
			))
		}
	}

	// 执行发送
	return yak.Send()
}

// -------------------------------------------------------------------
// SMTP LOGIN 认证实现
// -------------------------------------------------------------------

// 确保 smtpLoginAuth 实现了 smtp.Auth 接口
var _ smtp.Auth = (*smtpLoginAuth)(nil)

// smtpLoginAuth 定义了一个实现 LOGIN 认证机制的 AUTH
//
// AUTH LOGIN 已过时[1]，但某些邮件服务（如 Outlook）仍需要它[2]
//
// 注意！
// 此认证仅在连接使用 TLS 或连接到 localhost 时才会发送凭据。
// 否则认证将失败并返回错误，而不会发送凭据。
//
// [1]: https://github.com/golang/go/issues/40817
// [2]: https://support.microsoft.com/en-us/office/outlook-com-no-longer-supports-auth-plain-authentication-07f7d5e9-1697-465f-84d2-4513d4ff0145
type smtpLoginAuth struct {
	username, password string
}

// Start 初始化与服务器的认证
// 实现了 smtp.Auth 接口
func (a *smtpLoginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	// 必须使用 TLS，或者是 localhost 服务器
	// 注意：如果 TLS 不为 true，则不能信任 ServerInfo 中的任何内容
	// 特别是，服务器是否声明支持 LOGIN 认证并不重要
	// 那可能只是攻击者说"没事，你可以信任我的密码"
	if !server.TLS && !isLocalhost(server.Name) {
		return "", nil, errors.New("未加密连接")
	}

	return "LOGIN", nil, nil
}

// Next 通过向服务器提供请求的数据来"继续"认证过程
// 实现了 smtp.Auth 接口
func (a *smtpLoginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch strings.ToLower(string(fromServer)) {
		case "username:":
			return []byte(a.username), nil
		case "password:":
			return []byte(a.password), nil
		}
	}

	return nil, nil
}

// isLocalhost 检查服务器名称是否为本地主机
func isLocalhost(name string) bool {
	return name == "localhost" || name == "127.0.0.1" || name == "::1"
}
