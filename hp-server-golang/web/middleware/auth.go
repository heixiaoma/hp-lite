package middleware

import (
	"hp-server-lib/log"
	"hp-server-lib/util"
	"net/http"
	"strconv"
)

// 认证中间件 - 从token中提取用户ID并设置到请求头
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从请求头获取token
		token := r.Header.Get("token")
		log.Infof("AuthMiddleware: token=%s", token)
		if token != "" {
			// 解析token获取用户ID
			userId, role, timestamp, err := util.DecodeToken(token)
			log.Infof("AuthMiddleware: userId=%d, role=%s, timestamp=%d, err=%v", userId, role, timestamp, err)
			if err == nil && userId > 0 {
				// 将用户ID设置到请求头，供后续处理器使用
				r.Header.Set("X-User-ID", strconv.Itoa(userId))
				log.Infof("AuthMiddleware: Set X-User-ID=%d", userId)
			} else if err == nil && userId == -1 {
				// 管理员用户，也设置ID
				r.Header.Set("X-User-ID", strconv.Itoa(userId))
				log.Infof("AuthMiddleware: Set X-User-ID=%d (admin)", userId)
			}
		}
		// 继续处理请求
		next.ServeHTTP(w, r)
	}
}
