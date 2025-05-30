package util

import (
	"encoding/base32"
	"strings"
)

// 自定义小写 Base32 字母表
const lowerCaseBase32Alphabet = "abcdefghijklmnopqrstuvwxyz234567"

// 自定义编码实例
var lowerCaseEncoder = base32.NewEncoding(lowerCaseBase32Alphabet).WithPadding(base32.NoPadding)

// EncodeToLowerCaseBase32 将字符串编码为小写 Base32 格式
func EncodeToLowerCaseBase32(input string) string {
	return lowerCaseEncoder.EncodeToString([]byte(input))
}

// DecodeFromLowerCaseBase32 将小写 Base32 格式字符串解码为原始字符串
func DecodeFromLowerCaseBase32(encoded string) string {
	// 确保输入是小写
	encoded = strings.ToLower(encoded)

	decoded, err := lowerCaseEncoder.DecodeString(encoded)
	if err != nil {
		return ""
	}

	return string(decoded)
}
