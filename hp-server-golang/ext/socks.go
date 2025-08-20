package ext

import (
	"github.com/armon/go-socks5"
	"log"
	"net"
)

type SocksServer struct {
	listener net.Listener
	server   *socks5.Server
	port     string
	username string
	password string
}

// 创建 SOCKS5 服务器实例
func NewSocks(port string, username, password string) *SocksServer {
	return &SocksServer{
		port:     port,
		username: username,
		password: password,
	}
}

// 启动 SOCKS5 代理
func (s *SocksServer) Start(close func()) bool {
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
		log.Printf("创建 SOCKS5 服务器失败: %w", err)
		return false
	}
	s.server = server

	// 监听端口
	listener, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		log.Printf("监听端口失败: %w", err)
		return false
	}
	s.listener = listener

	// 启动服务（非阻塞）
	go func() {
		log.Printf("SOCKS5 代理已启动: 127.0.0.1:%s", s.port)
		if err := server.Serve(listener); err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "use of closed network connection" {
				// 正常关闭
				return
			}
			log.Printf("SOCKS5 服务错误: %v", err)
		}
		defer close()
	}()

	return true
}

// 停止 SOCKS5 代理
func (s *SocksServer) Stop() {
	if s.listener != nil {
		_ = s.listener.Close()
		log.Printf("SOCKS5 代理已停止: 127.0.0.1:%s", s.port)
	}
}
