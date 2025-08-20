package ext

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/armon/go-socks5"
)

type SocksServer struct {
	listener net.Listener
	server   *socks5.Server
	port     int
	username string
	password string
}

// 创建 SOCKS5 服务器实例
func NewSocks(port int, username, password string) *SocksServer {
	return &SocksServer{
		port:     port,
		username: username,
		password: password,
	}
}

// 启动 SOCKS5 代理
func (s *SocksServer) Start() error {
	config := &socks5.Config{}

	// 配置认证
	if s.username != "" && s.password != "" {
		cred := socks5.StaticCredentials{s.username: s.password}
		config.AuthMethods = []socks5.Authenticator{
			socks5.UserPassAuthenticator{Credentials: cred},
		}
	} else {
		config.AuthMethods = []socks5.Authenticator{
			socks5.NoAuthAuthenticator{},
		}
	}

	// 创建服务器
	server, err := socks5.New(config)
	if err != nil {
		return fmt.Errorf("创建 SOCKS5 服务器失败: %w", err)
	}
	s.server = server

	// 监听端口
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(s.port))
	if err != nil {
		return fmt.Errorf("监听端口失败: %w", err)
	}
	s.listener = listener

	// 启动服务（非阻塞）
	go func() {
		log.Printf("SOCKS5 代理已启动: 127.0.0.1:%d", s.port)
		if err := server.Serve(listener); err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "use of closed network connection" {
				// 正常关闭
				return
			}
			log.Printf("SOCKS5 服务错误: %v", err)
		}
	}()

	return nil
}

// 停止 SOCKS5 代理
func (s *SocksServer) Stop() {
	if s.listener != nil {
		_ = s.listener.Close()
		log.Printf("SOCKS5 代理已停止: 127.0.0.1:%d", s.port)
	}
}
