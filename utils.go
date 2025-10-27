package gomailer

import (
    mathRand "math/rand"
    "time"
)

const defaultRandomAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// pseudorandomString 生成指定长度的伪随机字符串
// 
// 生成的字符串匹配 [A-Za-z0-9]+ 模式，对 URL 编码透明
//
// 注意：此函数生成的是伪随机字符串，不适合用于安全敏感的场景
// 如果需要加密安全的随机字符串，请使用 crypto/rand 包
//
// 参数:
//   - length: 要生成的字符串长度
// 返回:
//   - string: 生成的随机字符串
func pseudorandomString(length int) string {
	return pseudorandomStringWithAlphabet(length, defaultRandomAlphabet)
}

// pseudorandomStringWithAlphabet 使用指定的字符集生成伪随机字符串
//
// 参数:
//   - length: 要生成的字符串长度
//   - alphabet: 可用字符集
// 返回:
//   - string: 生成的随机字符串
func pseudorandomStringWithAlphabet(length int, alphabet string) string {
    b := make([]byte, length)
    max := len(alphabet)

    for i := range b {
        b[i] = alphabet[mathRand.Intn(max)]
    }

    return string(b)
}

// 初始化伪随机种子，避免跨进程重复序列
func init() {
    mathRand.Seed(time.Now().UnixNano())
}
