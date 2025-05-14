package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

var aes_key = "hplite_token_key"

// AES 加密
func aesEncrypt(plainText, key []byte) (string, error) {
	// 生成 AES 块加密器
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 生成随机的 IV（初始化向量）
	ciphertext := make([]byte, aes.BlockSize+len(plainText))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// 创建加密模式
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plainText)

	// 返回 Base64 编码的加密数据
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// AES 解密
func aesDecrypt(cipherTextBase64, key []byte) (string, error) {
	// 解码 Base64
	cipherText, err := base64.URLEncoding.DecodeString(string(cipherTextBase64))
	if err != nil {
		return "", err
	}

	// 生成 AES 块加密器
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 提取 IV 和加密数据
	if len(cipherText) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	// 创建解密模式
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	// 返回解密后的明文
	return string(cipherText), nil
}

// 生成 Token（加密）
func GenerateToken(userId, role string) (string, error) {
	// 获取当前时间戳（毫秒）
	timestamp := time.Now().UnixMilli()

	// 拼接 userId + role + timestamp
	plainText := fmt.Sprintf("%s|%s|%d", userId, role, timestamp)

	// AES 加密并生成 Base64 编码的 Token
	return aesEncrypt([]byte(plainText), []byte(aes_key))
}

// 解密 Token
func DecodeToken(token string) (int, string, int64, error) {
	defer func() {
		if err := recover(); err != nil {
			// 捕获异常并记录日志
			log.Printf("解析Token错误: %v\n栈情况: %s", err, string(debug.Stack()))
		}
	}()

	// 解密 Base64 编码的 Token
	decodedText, err := aesDecrypt([]byte(token), []byte(aes_key))
	if err != nil {
		return 0, "", 0, err
	}
	parts := strings.Split(decodedText, "|")
	num, err := strconv.Atoi(parts[0])
	num2, err := strconv.ParseInt(parts[2], 10, 64)
	// 返回解密后的明文
	return num, parts[1], num2, nil
}
