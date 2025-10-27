package gomailer

import (
    "bytes"
    "errors"
    "mime"
    "net/http"
    "os/exec"
    "strings"
)

// 确保 Sendmail 实现了 Mailer 接口
var _ Mailer = (*Sendmail)(nil)

// Sendmail 实现了 Mailer 接口，定义了一个通过 "sendmail" *nix 命令发送邮件的客户端
//
// Sendmail 是一个在 Unix/Linux 系统上常用的邮件传输代理(MTA)
// 此客户端通过调用系统的 sendmail 命令来发送邮件
//
// 注意：此客户端通常仅推荐用于开发和测试环境
// 在生产环境中，建议使用 SMTPClient 以获得更好的控制和错误处理
type Sendmail struct {
	// onSend 发送钩子，允许在发送前后执行自定义逻辑
	onSend *Hook[*SendEvent]
}

// OnSend 实现 SendInterceptor 接口
// 返回发送钩子，允许用户在邮件发送前后添加自定义处理逻辑
//
// 示例:
//   client := &Sendmail{}
//   client.OnSend().BindFunc(func(e *SendEvent) error {
//       fmt.Println("准备发送邮件:", e.Message.Subject)
//       return e.Next()
//   })
func (c *Sendmail) OnSend() *Hook[*SendEvent] {
	if c.onSend == nil {
		c.onSend = &Hook[*SendEvent]{}
	}
	return c.onSend
}

// Send 实现 Mailer 接口
// 通过 sendmail 命令发送邮件
//
// 参数:
//   - m: 要发送的邮件消息
// 返回:
//   - error: 发送失败时返回错误，成功返回 nil
//
// 注意事项:
//   - 仅支持发送到 To 字段的收件人（不支持 Cc 和 Bcc）
//   - 不支持附件
//   - 优先发送 HTML 内容，如果没有 HTML 则发送纯文本
func (c *Sendmail) Send(m *Message) error {
	if c.onSend != nil {
		return c.onSend.Trigger(&SendEvent{Message: m}, func(e *SendEvent) error {
			return c.send(e.Message)
		})
	}

	return c.send(m)
}

// send 内部发送方法，执行实际的 sendmail 调用
func (c *Sendmail) send(m *Message) error {
    // 基础输入校验
    if m == nil {
        return errors.New("message is nil")
    }
    if m.From.Address == "" {
        return errors.New("from address is required")
    }
    if len(m.To) == 0 {
        return errors.New("at least one recipient in To is required")
    }

    // 提取收件人邮箱地址（不包含姓名）
    toAddresses := addressesToStrings(m.To, false)

	// 构建邮件头部
    headers := make(http.Header)
    headers.Set("Subject", mime.QEncoding.Encode("utf-8", m.Subject))
    headers.Set("From", m.From.String())
    // 根据正文选择合适的 Content-Type
    if m.HTML != "" {
        headers.Set("Content-Type", "text/html; charset=UTF-8")
    } else {
        headers.Set("Content-Type", "text/plain; charset=UTF-8")
    }
    headers.Set("To", strings.Join(toAddresses, ","))

	// 查找 sendmail 可执行文件路径
	cmdPath, err := findSendmailPath()
	if err != nil {
		return err
	}

	// 构建邮件内容
	var buffer bytes.Buffer

	// 写入邮件头部
	if err := headers.Write(&buffer); err != nil {
		return err
	}

	// 添加空行分隔头部和正文
	if _, err := buffer.Write([]byte("\r\n")); err != nil {
		return err
	}

    // 写入邮件正文（优先使用 HTML），确保至少有一个正文
    if m.HTML != "" {
        if _, err := buffer.Write([]byte(m.HTML)); err != nil {
            return err
        }
    } else if m.Text != "" {
        if _, err := buffer.Write([]byte(m.Text)); err != nil {
            return err
        }
    } else {
        // 回退一个最小正文，避免空 body 导致部分 MTA 拒收
        if _, err := buffer.Write([]byte("(empty body)")); err != nil {
            return err
        }
    }

    // 执行 sendmail 命令：以独立参数传递收件人
    // 参考：大多数 sendmail 兼容实现期望每个收件人为单独参数
    sendmail := exec.Command(cmdPath, toAddresses...)
    sendmail.Stdin = &buffer

    return sendmail.Run()
}

// findSendmailPath 查找系统中 sendmail 可执行文件的路径
//
// 返回:
//   - string: sendmail 可执行文件的完整路径
//   - error: 如果找不到 sendmail，返回错误
//
// 搜索顺序:
//   1. /usr/sbin/sendmail（大多数 Linux 发行版）
//   2. /usr/bin/sendmail（某些 Unix 系统）
//   3. sendmail（在 PATH 环境变量中查找）
func findSendmailPath() (string, error) {
	options := []string{
		"/usr/sbin/sendmail",
		"/usr/bin/sendmail",
		"sendmail",
	}

	for _, option := range options {
		path, err := exec.LookPath(option)
		if err == nil {
			return path, err
		}
	}

	return "", errors.New("无法找到 sendmail 可执行文件路径")
}
