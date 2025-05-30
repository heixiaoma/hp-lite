package main

import (
	"encoding/base32"
	"fmt"
	"strings"
)

func main() {
	// 测试用例
	testCases := []string{
		"Hello, World!",
		"Golang Base32 Encoding",
		"1234567890",
		"!@#$%^&*()",
		"你好，世界！",
		"A very long string that needs to be encoded using lowercase characters only.",
	}

	for _, test := range testCases {
		// 编码
		encoded := EncodeToLowerCaseBase32(test)

		// 解码
		decoded, err := DecodeFromLowerCaseBase32(encoded)
		if err != nil {
			fmt.Printf("解码错误: %v\n", err)
			continue
		}

		// 输出结果
		fmt.Printf("原始字符串: %s\n", test)
		fmt.Printf("小写 Base32: %s\n", encoded)
		fmt.Printf("解码结果:   %s\n", decoded)
		fmt.Printf("是否匹配:   %v\n", test == decoded)
		fmt.Printf("原始长度:   %d 字节\n", len(test))
		fmt.Printf("编码后长度: %d 字节\n", len(encoded))
		fmt.Println("-------------------")
	}
}
