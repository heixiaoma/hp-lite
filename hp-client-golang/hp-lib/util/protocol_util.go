package util

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// ProtocolInfo 解析地址，返回错误、协议、地址（域名/IP/套接字路径）、端口
// 注意：1. 域名不解析为IP，原样返回；2. unix协议返回套接字路径，端口固定为0
func ProtocolInfo(address string) (error, string, string, int) {
	var protocol string
	var target string // 协议对应的目标（主机:端口 或 套接字路径）

	// 1. 提取协议类型（优先检查unix，避免与其他协议冲突）
	switch {
	case strings.HasPrefix(address, "unix://"):
		protocol = "unix"
		target = strings.TrimPrefix(address, "unix://")
		if target == "" {
			return errors.New("unix协议地址格式错误，需指定套接字路径（如 unix:///tmp/sock）"), "", "", 0
		}
	case strings.HasPrefix(address, "http://"):
		protocol = "http"
		target = strings.TrimPrefix(address, "http://")
	case strings.HasPrefix(address, "https://"):
		protocol = "https"
		target = strings.TrimPrefix(address, "https://")
	case strings.HasPrefix(address, "tcp://"):
		protocol = "tcp"
		target = strings.TrimPrefix(address, "tcp://")
	case strings.HasPrefix(address, "udp://"):
		protocol = "udp"
		target = strings.TrimPrefix(address, "udp://")
	case strings.HasPrefix(address, "socks5://"):
		protocol = "socks5"
		// 移除协议头后，先剥离认证信息（仅保留服务器地址部分用于主解析）
		target = strings.TrimPrefix(address, "socks5://")
		if atIdx := strings.Index(target, "@"); atIdx != -1 {
			target = target[atIdx+1:] // 从@后面开始截取（去掉 用户名:密码@ 部分）
		}
	default:
		return errors.New("不支持的协议格式"), "", "", 0
	}

	// 2. 特殊处理unix协议（无端口，地址为文件路径）
	if protocol == "unix" {
		return nil, protocol, target, 0
	}

	// 3. 处理HTTP/HTTPS中的路径、参数等（仅保留主机和端口）
	if protocol == "http" || protocol == "https" {
		fullURL := protocol + "://" + target
		u, err := url.Parse(fullURL)
		if err != nil {
			return fmt.Errorf("URL解析失败: %w", err), "", "", 0
		}
		target = u.Host // 提取 域名:端口 或 IP:端口
	}

	// 4. 拆分主机（域名或IP）和端口
	host, portStr, err := net.SplitHostPort(target)
	if err != nil {
		// 无显式端口时使用协议默认端口
		host = target
		switch protocol {
		case "http":
			portStr = "80"
		case "https":
			portStr = "443"
		case "socks5":
			portStr = "1080"
		default: // TCP/UDP必须显式指定端口
			return errors.New("tcp/udp协议必须指定端口（如 tcp://ip:port 或 tcp://域名:port）"), "", "", 0
		}
	}

	// 5. 端口转换为int类型
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("端口格式错误: %w", err), "", "", 0
	}

	// 6. 不解析域名，直接返回主机（域名或IP原样保留）
	return nil, protocol, host, port
}

// ParseSocks5Auth 从socks5地址中解析用户名和密码（仅支持 socks5://user:pass@host:port 格式）
// 若地址不包含认证信息或格式错误，返回空字符串和错误
func ParseSocks5Auth(address string) (username, password string, err error) {
	// 先检查是否为socks5协议
	if !strings.HasPrefix(address, "socks5://") {
		return "", "", errors.New("不是socks5协议地址")
	}

	// 移除协议头后解析认证信息
	target := strings.TrimPrefix(address, "socks5://")
	atIdx := strings.Index(target, "@")
	if atIdx == -1 {
		// 无@符号，说明没有认证信息
		return "", "", nil // 也可返回错误，根据需求调整
	}

	// 截取@前面的认证部分（user:pass）
	authPart := target[:atIdx]
	up := strings.SplitN(authPart, ":", 2)
	if len(up) < 1 {
		return "", "", errors.New("socks5认证信息格式错误（缺少用户名）")
	}

	username = up[0]
	if len(up) == 2 {
		password = up[1]
	} else {
		password = "" // 允许空密码
	}

	return username, password, nil
}
