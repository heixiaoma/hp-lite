package util

import "net"

// IsIPInCIDR 判断 IP 是否属于指定的 CIDR 网段
func IsIPInCIDR(ipStr, cidrStr string) bool {
	// 解析 IP 地址
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false // 无效的 IP 地址
	}

	// 解析 CIDR 网段
	_, ipnet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return false // 无效的 CIDR 格式
	}
	// 判断 IP 是否在网段内
	return ipnet.Contains(ip)
}

// GetClientIP 从 net.Conn 获取客户端 IP 地址
func GetClientIP(conn net.Conn) string {
	// 获取远程地址（格式：IP:端口）
	remoteAddr := conn.RemoteAddr().String()

	// 分割主机和端口
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr // 无法分割时返回原始地址
	}

	return host
}

// GetClientIPFromUDP 获取 UDP 数据报发送方的 IP 地址
func GetClientIPFromUDP(raddr net.Addr) string {
	// 类型断言确保是 UDP 地址
	if udpAddr, ok := raddr.(*net.UDPAddr); ok {
		return udpAddr.IP.String()
	}
	return raddr.String()
}
