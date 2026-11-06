package uuidTool

import "fmt"

// ase64KeyChars 是标准 Base64 的字符表。
// 这里并不是按字节做标准 Base64 编码，而是把 6-bit 数值映射成字符。
const ase64KeyChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

// hexCharValue 把单个十六进制字符转成 0~15 的数值。
// 例如 '0'->0, 'a'->10, 'F'->15。
func hexCharValue(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	default:
		panic(fmt.Sprintf("invalid hex character: %q", c))
	}
}

// hexByte 把 0~15 的数值转成一个小写十六进制字符。
func hexByte(n int) byte {
	return "0123456789abcdef"[n]
}
