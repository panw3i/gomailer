package gomailer

import (
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// 匹配连续空白字符的正则表达式
var whitespaceRegex = regexp.MustCompile(`\s+`)

// 需要跳过的 HTML 标签（不包含在转换后的文本中）
var tagsToSkip = []string{
	"style",    // 样式定义
	"script",   // JavaScript 脚本
	"iframe",   // 内联框架
	"applet",   // Java 小程序
	"object",   // 嵌入对象
	"svg",      // SVG 矢量图
	"img",      // 图片
	"button",   // 按钮
	"form",     // 表单
	"textarea", // 文本域
	"input",    // 输入框
	"select",   // 下拉选择框
	"option",   // 选择选项
	"template", // 模板
}

// 内联标签（不会触发换行）
var inlineTags = []string{
	"a",      // 链接
	"span",   // 行内容器
	"small",  // 小号文本
	"strike", // 删除线
	"strong", // 加粗强调
	"sub",    // 下标
	"sup",    // 上标
	"em",     // 斜体强调
	"b",      // 粗体
	"u",      // 下划线
	"i",      // 斜体
}

// html2Text 是一个非常基础的 HTML 到纯文本的自动转换器
// 用于在没有提供纯文本版本时，从 HTML 邮件正文生成纯文本版本
//
// 参数:
//   - htmlDocument: HTML 文档字符串
//
// 返回:
//   - string: 转换后的纯文本
//   - error: 解析失败时返回错误
//
// 注意事项:
//   - 此方法不检查 HTML 文档的正确性
//   - 链接将转换为 "[文本](url)" 格式
//   - 列表项 (<li>) 以 "- " 为前缀
//   - 缩进会被去除（包括制表符和空格）
//   - 尾随空格会被保留
//   - 多个连续换行符会被合并为一个，除非使用了多个 <br> 标签
func html2Text(htmlDocument string) (string, error) {
	// 解析 HTML 文档
	doc, err := html.Parse(strings.NewReader(htmlDocument))
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	var canAddNewLine bool // 标记是否可以添加新行

	// 递归遍历 HTML 节点树
	// 参考: https://pkg.go.dev/golang.org/x/net/html#Parse
	var f func(*html.Node, *strings.Builder)
	f = func(n *html.Node, activeBuilder *strings.Builder) {
		// 检查是否为链接节点
		isLink := n.Type == html.ElementNode && n.Data == "a"

		if isLink {
			// 链接使用单独的 builder 来收集链接文本
			var linkBuilder strings.Builder
			activeBuilder = &linkBuilder
		} else if activeBuilder == nil {
			activeBuilder = &builder
		}

		switch n.Type {
		case html.TextNode:
			// 处理文本节点
			// 将多个连续空白字符替换为单个空格
			txt := whitespaceRegex.ReplaceAllString(n.Data, " ")

			// 如果前一个节点有换行，可以安全地去除缩进
			if !canAddNewLine {
				txt = strings.TrimLeft(txt, " ")
			}

			if txt != "" {
				activeBuilder.WriteString(txt)
				canAddNewLine = true
			}

		case html.ElementNode:
			// 处理元素节点
			if n.Data == "br" {
				// <br> 标签始终写入换行
				activeBuilder.WriteString("\r\n")
				canAddNewLine = false
			} else if canAddNewLine && !existInSlice(n.Data, inlineTags) {
				// 块级元素添加换行
				activeBuilder.WriteString("\r\n")
				canAddNewLine = false
			}

			// 为列表项添加前缀
			if n.Data == "li" {
				activeBuilder.WriteString("- ")
			}
		}

		// 递归处理子节点
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			// 跳过不需要的标签
			if c.Type != html.ElementNode || !existInSlice(c.Data, tagsToSkip) {
				f(c, activeBuilder)
			}
		}

		// 格式化链接为 [label](href)
		if isLink {
			linkTxt := strings.TrimSpace(activeBuilder.String())
			if linkTxt == "" {
				linkTxt = "LINK"
			}

			builder.WriteString("[")
			builder.WriteString(linkTxt)
			builder.WriteString("]")

			// 提取链接的 href 属性
			for _, a := range n.Attr {
				if a.Key == "href" {
					if a.Val != "" {
						builder.WriteString("(")
						builder.WriteString(a.Val)
						builder.WriteString(")")
					}
					break
				}
			}

			activeBuilder.Reset()
		}
	}

	// 开始转换
	f(doc, &builder)

	return strings.TrimSpace(builder.String()), nil
}

// existInSlice 检查字符串是否存在于切片中
func existInSlice(item string, list []string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}

