package web

import (
	"embed"
	"fmt"
	"hp-server-lib/log"
	"hp-server-lib/web/controller"
	"net/http"
	"runtime/debug"
	"strconv"
)

//go:embed static
var content embed.FS

// 全局异常拦截器中间件
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		defer func() {
			if err := recover(); err != nil {
				// 捕获异常并记录日志
				log.Errorf("服务器错误: %v\n栈情况: %s", err, string(debug.Stack()))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, `{"error": "服务器错误", "message": "%v"}`, err)
			}
		}()
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func StartWebServer(port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", controller.StaticController{Content: content}.Static)

	mux.HandleFunc("/user/login", controller.LoginController{}.LoginHandler)
	clientUserController := controller.ClientUserController{}
	mux.HandleFunc("/client/user/saveUser", clientUserController.Add)
	mux.HandleFunc("/client/user/list", clientUserController.List)
	mux.HandleFunc("/client/user/removeUser", clientUserController.Del)

	deviceController := controller.DeviceController{}
	mux.HandleFunc("/client/device/list", deviceController.List)
	mux.HandleFunc("/client/device/add", deviceController.Add)
	mux.HandleFunc("/client/device/update", deviceController.Update)
	mux.HandleFunc("/client/device/remove", deviceController.Del)
	mux.HandleFunc("/client/device/stop", deviceController.Stop)

	configController := controller.ConfigController{}
	mux.HandleFunc("/client/config/getDeviceKey", configController.GetDeviceKey)
	mux.HandleFunc("/client/config/getConfigList", configController.GetConfigList)
	mux.HandleFunc("/client/config/removeConfig", configController.RemoveConfig)
	mux.HandleFunc("/client/config/refConfig", configController.RefConfig)
	mux.HandleFunc("/client/config/changeStatus", configController.ChangeStatus)
	mux.HandleFunc("/client/config/addConfig", configController.Add)
	mux.HandleFunc("/client/config/keyword", configController.Keyword)

	monitorController := controller.MonitorController{}
	mux.HandleFunc("/client/monitor/list", monitorController.List)
	mux.HandleFunc("/client/monitor/detail", monitorController.Detail)

	domainController := controller.DomainController{}
	mux.HandleFunc("/client/domain/list", domainController.GetDomainList)
	mux.HandleFunc("/client/domain/remove", domainController.RemoveDomain)
	mux.HandleFunc("/client/domain/add", domainController.Add)
	mux.HandleFunc("/client/domain/gen", domainController.Gen)
	mux.HandleFunc("/client/domain/query", domainController.Query)

	wafController := controller.WafController{}
	mux.HandleFunc("/client/waf/save", wafController.Add)
	mux.HandleFunc("/client/waf/list", wafController.List)
	mux.HandleFunc("/client/waf/remove", wafController.Del)

	reverseController := controller.ReverseController{}
	mux.HandleFunc("/client/reverse/save", reverseController.Add)
	mux.HandleFunc("/client/reverse/list", reverseController.List)
	mux.HandleFunc("/client/reverse/remove", reverseController.Del)

	forwardController := controller.ForwardController{}
	mux.HandleFunc("/client/forward/save", forwardController.Add)
	mux.HandleFunc("/client/forward/list", forwardController.List)
	mux.HandleFunc("/client/forward/remove", forwardController.Del)

	giscusController := controller.GiscusController{}
	mux.HandleFunc("/client/giscus/token", giscusController.Token)

	muxWithRecovery := recoveryMiddleware(mux)
	log.Error(http.ListenAndServe(":"+strconv.Itoa(port), muxWithRecovery))
}
