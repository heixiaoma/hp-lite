package web

import (
	"fmt"
	"hp-server-lib/web/controller"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
)

// 全局异常拦截器中间件
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
		}
		defer func() {
			if err := recover(); err != nil {
				// 捕获异常并记录日志
				log.Printf("Recovered from panic: %v\nStackTrace: %s", err, string(debug.Stack()))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, `{"error": "Internal Server Error", "message": "%v"}`, err)
			}
		}()
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func StartWebServer(port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user/login", controller.LoginController{}.LoginHandler)
	clientUserController := controller.ClientUserController{}
	mux.HandleFunc("/client/user/saveUser", clientUserController.Add)
	mux.HandleFunc("/client/user/list", clientUserController.List)
	mux.HandleFunc("/client/user/removeUser", clientUserController.Del)
	// 使用全局异常处理器
	muxWithRecovery := recoveryMiddleware(mux)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), muxWithRecovery))
}
