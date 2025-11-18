package ext

import (
	"hp-server-lib/ext/socks5"
	"hp-server-lib/log"
	"net"
	"time"
)

type SocksServer struct {
	socks5.Server
	listener net.Listener
	port     string
	username string
	password string
}

// 创建 SOCKS5 服务器实例
func NewSocks(port string, username, password string) *SocksServer {
	s := &SocksServer{
		port:     port,
		username: username,
		password: password,
	}
	s.ListenBindReuseTimeout = time.Second / 2
	if s.username != "" && s.password != "" {
		s.Authentication = socks5.UserAuth(s.username, s.password)
	}
	return s
}

// 启动 SOCKS5 代理
func (s *SocksServer) Start(close func()) bool {
	listener, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		log.Errorf("SOCKS5 服务错误: %v", err)
		return false
	}
	s.listener = listener
	go func() {
		defer close()
		if err := s.Serve(s.listener); err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "use of closed network connection" {
				// 正常关闭
				return
			}
			log.Errorf("SOCKS5 服务错误: %v", err)
		}
	}()
	return true
}

// 停止 SOCKS5 代理
func (s *SocksServer) Stop() {
	if s.listener != nil {
		_ = s.listener.Close()
		log.Infof("SOCKS5 代理已停止:%s", s.port)
	}
}
