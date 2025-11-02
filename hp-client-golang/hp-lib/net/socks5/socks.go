package net

import (
	"net"

	"github.com/armon/go-socks5"
)

type SocksServer struct {
	listener net.Listener
	server   *socks5.Server
	port     string
	username string
	password string
	callMsg  func(message string)
}

// 创建 SOCKS5 服务器实例
func NewSocks(port string, username, password string, callMsg func(message string)) *SocksServer {
	return &SocksServer{
		port:     port,
		username: username,
		password: password,
		callMsg:  callMsg,
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
		s.callMsg("创建 SOCKS5 服务器失败: " + err.Error())
		return false
	}
	s.server = server

	// 监听端口
	listener, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		s.callMsg("监听端口失败: " + err.Error())
		return false
	}
	s.listener = listener

	// 启动服务（非阻塞）
	go func() {
		s.callMsg("SOCKS5 代理已启动 :" + s.port)
		if err := server.Serve(listener); err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "use of closed network connection" {
				// 正常关闭
				return
			}
			s.callMsg("SOCKS5 服务错误 :" + err.Error())
		}
		defer close()
	}()

	return true
}

// 停止 SOCKS5 代理
func (s *SocksServer) Stop() {
	if s.listener != nil {
		_ = s.listener.Close()
		s.callMsg("SOCKS5 代理已停止:" + s.port)
	}
}
