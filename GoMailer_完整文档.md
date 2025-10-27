# GoMailer 使用说明

GoMailer 是一个简单易用的 Go 语言邮件发送库，支持 SMTP 和 Sendmail 两种发送方式。

## 安装

```bash
go get github.com/panw3i/gomailer
```

## 快速开始

### SMTP 发送邮件

```go
package main

import (
    "log"
    "net/mail"

    "github.com/panw3i/gomailer"
)

func main() {
    client := &gomailer.SMTPClient{
        Host:     "smtp.gmail.com",
        Port:     587,
        Username: "your-email@gmail.com",
        Password: "your-app-password",
        TLS:      true,
    }

    message := &gomailer.Message{
        From: mail.Address{
            Name:    "发件人",
            Address: "sender@example.com",
        },
        To: []mail.Address{
            {Address: "recipient@example.com"},
        },
        Subject: "测试邮件",
        HTML:    "<h1>你好！</h1><p>这是一封测试邮件。</p>",
    }

    if err := client.Send(message); err != nil {
        log.Fatal("发送失败:", err)
    }

    log.Println("邮件发送成功！")
}
```

### Sendmail 发送邮件

```go
package main

import (
    "log"
    "net/mail"

    "github.com/panw3i/gomailer"
)

func main() {
    client := &gomailer.Sendmail{}

    message := &gomailer.Message{
        From: mail.Address{
            Name:    "发件人",
            Address: "sender@example.com",
        },
        To: []mail.Address{
            {Address: "recipient@example.com"},
        },
        Subject: "测试邮件",
        HTML:    "<h1>你好！</h1><p>这是一封测试邮件。</p>",
    }

    if err := client.Send(message); err != nil {
        log.Fatal("发送失败:", err)
    }

    log.Println("邮件发送成功！")
}
```

## 邮件结构

```go
type Message struct {
    From              mail.Address         // 发件人
    To                []mail.Address       // 收件人列表
    Bcc               []mail.Address       // 密送列表
    Cc                []mail.Address       // 抄送列表
    Subject           string               // 邮件主题
    HTML              string               // HTML 正文
    Text              string               // 纯文本正文
    Headers           map[string]string    // 自定义邮件头
    Attachments       map[string]io.Reader // 普通附件
    InlineAttachments map[string]io.Reader // 内联附件
}
```

## SMTP 客户端配置

```go
type SMTPClient struct {
    Host       string  // SMTP 服务器地址
    Port       int     // SMTP 端口（通常为 25、465 或 587）
    Username   string  // 认证用户名
    Password   string  // 认证密码
    TLS        bool    // 是否使用 TLS 加密
    AuthMethod string  // 认证方法（PLAIN 或 LOGIN）
    LocalName  string  // 本地主机名（某些服务器需要）
}
```

### 常见 SMTP 配置

#### Gmail
```go
client := &gomailer.SMTPClient{
    Host:     "smtp.gmail.com",
    Port:     587,
    Username: "your-email@gmail.com",
    Password: "your-app-password", // 使用应用专用密码
    TLS:      true,
}
```

#### Outlook
```go
client := &gomailer.SMTPClient{
    Host:       "smtp-mail.outlook.com",
    Port:       587,
    Username:   "your-email@outlook.com",
    Password:   "your-password",
    TLS:        true,
    AuthMethod: gomailer.SMTPAuthLogin, // Outlook 需要 LOGIN 认证
}
```

#### Amazon SES
```go
client := &gomailer.SMTPClient{
    Host:     "email-smtp.ap-northeast-1.amazonaws.com",
    Port:     587,
    Username: os.Getenv("SES_SMTP_USER"),
    Password: os.Getenv("SES_SMTP_PASS"),
    TLS:      true,
}
```

#### QQ 邮箱
```go
client := &gomailer.SMTPClient{
    Host:     "smtp.qq.com",
    Port:     587,
    Username: "your-email@qq.com",
    Password: "your-authorization-code", // 使用授权码
    TLS:      true,
}
```

#### 163 邮箱
```go
client := &gomailer.SMTPClient{
    Host:     "smtp.163.com",
    Port:     25,
    Username: "your-email@163.com",
    Password: "your-authorization-code", // 使用授权码
    TLS:      false,
}
```

## 高级用法

### 发送带附件的邮件

```go
package main

import (
    "log"
    "net/mail"
    "os"

    "github.com/panw3i/gomailer"
)

func main() {
    client := &gomailer.SMTPClient{
        Host:     "smtp.gmail.com",
        Port:     587,
        Username: "your-email@gmail.com",
        Password: "your-app-password",
        TLS:      true,
    }

    file, err := os.Open("document.pdf")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    message := &gomailer.Message{
        From: mail.Address{
            Name:    "发件人",
            Address: "sender@example.com",
        },
        To: []mail.Address{
            {Address: "recipient@example.com"},
        },
        Subject: "带附件的邮件",
        HTML:    "<p>请查收附件。</p>",
        Attachments: map[string]io.Reader{
            "document.pdf": file, // 文件名 -> 文件读取器
        },
    }

    if err := client.Send(message); err != nil {
        log.Fatal(err)
    }

    log.Println("邮件发送成功！")
}
```

### 发送带内联图片的邮件

```go
package main

import (
    "log"
    "net/mail"
    "os"

    "github.com/panw3i/gomailer"
)

func main() {
    client := &gomailer.SMTPClient{
        Host:     "smtp.gmail.com",
        Port:     587,
        Username: "your-email@gmail.com",
        Password: "your-app-password",
        TLS:      true,
    }

    logo, err := os.Open("logo.png")
    if err != nil {
        log.Fatal(err)
    }
    defer logo.Close()

    message := &gomailer.Message{
        From: mail.Address{
            Name:    "发件人",
            Address: "sender@example.com",
        },
        To: []mail.Address{
            {Address: "recipient@example.com"},
        },
        Subject: "带图片的邮件",
        HTML:    `<h1>欢迎！</h1><img src="cid:logo.png" alt="Logo" />`,
        InlineAttachments: map[string]io.Reader{
            "logo.png": logo, // 在 HTML 中使用 cid:logo.png 引用
        },
    }

    if err := client.Send(message); err != nil {
        log.Fatal(err)
    }

    log.Println("邮件发送成功！")
}
```

### 使用钩子函数

```go
package main

import (
    "log"
    "net/mail"
    "time"

    "github.com/panw3i/gomailer"
)

func main() {
    client := &gomailer.SMTPClient{
        Host:     "smtp.gmail.com",
        Port:     587,
        Username: "your-email@gmail.com",
        Password: "your-app-password",
        TLS:      true,
    }

    // 添加发送前的钩子
    client.OnSend().BindFunc(func(e *gomailer.SendEvent) error {
        log.Printf("准备发送邮件: %s -> %s",
            e.Message.From.Address,
            e.Message.To[0].Address,
        )

        start := time.Now()

        // 继续执行发送
        err := e.Next()

        // 记录发送耗时
        duration := time.Since(start)
        if err != nil {
            log.Printf("发送失败 (耗时 %v): %v", duration, err)
        } else {
            log.Printf("发送成功 (耗时 %v)", duration)
        }

        return err
    })

    message := &gomailer.Message{
        From: mail.Address{
            Name:    "发件人",
            Address: "sender@example.com",
        },
        To: []mail.Address{
            {Address: "recipient@example.com"},
        },
        Subject: "测试邮件",
        HTML:    "<p>这是一封测试邮件。</p>",
    }

    if err := client.Send(message); err != nil {
        log.Fatal(err)
    }
}
```

### 发送给多个收件人

```go
message := &gomailer.Message{
    From: mail.Address{
        Name:    "发件人",
        Address: "sender@example.com",
    },
    To: []mail.Address{
        {Name: "张三", Address: "zhangsan@example.com"},
        {Name: "李四", Address: "lisi@example.com"},
    },
    Cc: []mail.Address{
        {Address: "manager@example.com"}, // 抄送
    },
    Bcc: []mail.Address{
        {Address: "admin@example.com"}, // 密送
    },
    Subject: "团队通知",
    HTML:    "<p>这是一封群发邮件。</p>",
}
```

### 自定义邮件头

```go
message := &gomailer.Message{
    From: mail.Address{
        Name:    "发件人",
        Address: "sender@example.com",
    },
    To: []mail.Address{
        {Address: "recipient@example.com"},
    },
    Subject: "自定义邮件头",
    HTML:    "<p>这是一封带自定义头的邮件。</p>",
    Headers: map[string]string{
        "X-Priority":       "1",           // 高优先级
        "X-Custom-Header":  "Custom Value", // 自定义头
        "Reply-To":         "reply@example.com",
    },
}
```

## 使用场景示例

### 用户注册验证邮件

```go
func SendVerificationEmail(userEmail, token string) error {
    client := &gomailer.SMTPClient{
        Host:     "smtp.example.com",
        Port:     587,
        Username: "noreply@example.com",
        Password: "password",
        TLS:      true,
    }

    verifyURL := fmt.Sprintf("https://example.com/verify?token=%s", token)

    message := &gomailer.Message{
        From: mail.Address{
            Name:    "Example 团队",
            Address: "noreply@example.com",
        },
        To: []mail.Address{
            {Address: userEmail},
        },
        Subject: "请验证您的邮箱",
        HTML: fmt.Sprintf(`
            <h2>欢迎注册 Example！</h2>
            <p>请点击下面的链接验证您的邮箱：</p>
            <p><a href="%s">验证邮箱</a></p>
            <p>如果您没有注册账号，请忽略此邮件。</p>
        `, verifyURL),
    }

    return client.Send(message)
}
```

### 密码重置邮件

```go
func SendPasswordResetEmail(userEmail, resetToken string) error {
    client := &gomailer.SMTPClient{
        Host:     "smtp.example.com",
        Port:     587,
        Username: "noreply@example.com",
        Password: "password",
        TLS:      true,
    }

    resetURL := fmt.Sprintf("https://example.com/reset-password?token=%s", resetToken)

    message := &gomailer.Message{
        From: mail.Address{
            Name:    "Example 团队",
            Address: "noreply@example.com",
        },
        To: []mail.Address{
            {Address: userEmail},
        },
        Subject: "重置您的密码",
        HTML: fmt.Sprintf(`
            <h2>密码重置请求</h2>
            <p>我们收到了重置您密码的请求。</p>
            <p>请点击下面的链接重置密码（链接在 1 小时内有效）：</p>
            <p><a href="%s">重置密码</a></p>
            <p>如果您没有请求重置密码，请忽略此邮件。</p>
        `, resetURL),
    }

    return client.Send(message)
}
```

## 注意事项

1. **Gmail 用户**：需要使用[应用专用密码](https://support.google.com/accounts/answer/185833)，而不是账号密码
2. **国内邮箱**：多数需要开启 SMTP 服务并使用授权码
3. **Sendmail**：仅推荐在开发环境使用，生产环境建议使用 SMTP
4. **TLS 连接**：推荐始终使用 TLS 加密连接以保护邮件内容和认证信息
5. **端口与 TLS**：
   - `465 + TLS=true` 使用隐式 TLS；
   - `587/25 + TLS=true` 使用 STARTTLS 升级；
   - `TLS=false` 为明文连接，不推荐。
6. **发送频率**：注意各邮件服务商的发送频率限制，避免被标记为垃圾邮件
7. **HTML 内容**：确保 HTML 邮件在各种邮件客户端中都能正常显示

## API 参考

### Mailer 接口

```go
type Mailer interface {
    Send(message *Message) error
}
```

### SMTPClient 方法

- `Send(message *Message) error` - 发送邮件
- `OnSend() *Hook[*SendEvent]` - 获取发送钩子

### Sendmail 方法

- `Send(message *Message) error` - 发送邮件
- `OnSend() *Hook[*SendEvent]` - 获取发送钩子

### Hook 方法

- `Bind(handler *Handler[T]) string` - 绑定处理器
- `BindFunc(fn func(e T) error) string` - 绑定处理器函数
- `Unbind(ids ...string)` - 解绑处理器
- `UnbindAll()` - 解绑所有处理器
- `Trigger(event T, oneOffHandlerFuncs ...func(T) error) error` - 触发钩子

## 获取帮助

如有问题或建议，请：
- 提交 [Issue](https://github.com/panw3i/gomailer/issues)
- 发送邮件至 your-email@example.com

---

**注意**：使用前请确保已正确配置 SMTP 服务器信息。