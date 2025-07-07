package tunnel

import "sync"

// ConnectionLimiter 连接限流器
type ConnectionLimiter struct {
	maxConnections int
	semaphore      chan struct{}
	mu             sync.RWMutex
	activeConns    int
}

// NewConnectionLimiter 创建新的连接限流器
func NewConnectionLimiter(max int) *ConnectionLimiter {
	return &ConnectionLimiter{
		maxConnections: max,
		semaphore:      make(chan struct{}, max),
	}
}

// Acquire 获取连接许可
func (l *ConnectionLimiter) Acquire() bool {
	select {
	case l.semaphore <- struct{}{}:
		l.mu.Lock()
		l.activeConns++
		l.mu.Unlock()
		return true
	default:
		return false
	}
}

// Release 释放连接许可
func (l *ConnectionLimiter) Release() {
	<-l.semaphore
	l.mu.Lock()
	l.activeConns--
	l.mu.Unlock()
}

// ActiveConnections 获取当前活跃连接数
func (l *ConnectionLimiter) ActiveConnections() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.activeConns
}

// MaxConnections 获取最大连接数限制
func (l *ConnectionLimiter) MaxConnections() int {
	return l.maxConnections
}
