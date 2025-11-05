package socks5

import (
	"log"
	"net"
	"os"
	"time"
)

type SocksServer struct {
	Server
	listener net.Listener
	port     string
	username string
	password string
	callMsg  func(message string)
}

// 创建 SOCKS5 服务器实例
func NewSocks(port string, username, password string, callMsg func(message string)) *SocksServer {
	s := &SocksServer{
		port:     port,
		username: username,
		password: password,
		callMsg:  callMsg,
	}
	s.Logger = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	s.ListenBindReuseTimeout = time.Second / 2
	if s.username != "" && s.password != "" {
		s.Authentication = UserAuth(s.username, s.password)
	}
	return s
}

// 启动 SOCKS5 代理
func (s *SocksServer) Start(close func()) bool {
	listener, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		s.callMsg("SOCKS5 服务错误: " + err.Error())
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
			s.callMsg("SOCKS5 服务错误: " + err.Error())
		}
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
