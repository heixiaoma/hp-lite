package web

import (
	"embed"
	"fmt"
	"hp-server-lib/log"
	"hp-server-lib/web/controller"
	"hp-server-lib/web/middleware"
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
	mux.HandleFunc("/client/user/saveUser", middleware.Admin(clientUserController.Add))
	mux.HandleFunc("/client/user/list", middleware.Admin(clientUserController.List))
	mux.HandleFunc("/client/user/removeUser", middleware.Admin(clientUserController.Del))

	deviceController := controller.DeviceController{}
	mux.HandleFunc("/client/device/list", middleware.Auth(deviceController.List))
	mux.HandleFunc("/client/device/add", middleware.Auth(deviceController.Add))
	mux.HandleFunc("/client/device/update", middleware.Auth(deviceController.Update))
	mux.HandleFunc("/client/device/remove", middleware.Auth(deviceController.Del))
	mux.HandleFunc("/client/device/stop", middleware.Auth(deviceController.Stop))

	configController := controller.ConfigController{}
	mux.HandleFunc("/client/config/getDeviceKey", middleware.Auth(configController.GetDeviceKey))
	mux.HandleFunc("/client/config/getConfigList", middleware.Auth(configController.GetConfigList))
	mux.HandleFunc("/client/config/removeConfig", middleware.Auth(configController.RemoveConfig))
	mux.HandleFunc("/client/config/refConfig", middleware.Auth(configController.RefConfig))
	mux.HandleFunc("/client/config/changeStatus", middleware.Auth(configController.ChangeStatus))
	mux.HandleFunc("/client/config/addConfig", middleware.Auth(configController.Add))
	mux.HandleFunc("/client/config/keyword", middleware.Auth(configController.Keyword))

	monitorController := controller.MonitorController{}
	mux.HandleFunc("/client/monitor/list", middleware.Auth(monitorController.List))
	mux.HandleFunc("/client/monitor/detail", middleware.Auth(monitorController.Detail))

	domainController := controller.DomainController{}
	mux.HandleFunc("/client/domain/list", middleware.Auth(domainController.GetDomainList))
	mux.HandleFunc("/client/domain/remove", middleware.Auth(domainController.RemoveDomain))
	mux.HandleFunc("/client/domain/add", middleware.Auth(domainController.Add))
	mux.HandleFunc("/client/domain/gen", middleware.Auth(domainController.Gen))
	mux.HandleFunc("/client/domain/query", middleware.Auth(domainController.Query))

	wafController := controller.WafController{}
	mux.HandleFunc("/client/waf/save", middleware.Auth(wafController.Add))
	mux.HandleFunc("/client/waf/list", middleware.Auth(wafController.List))
	mux.HandleFunc("/client/waf/remove", middleware.Auth(wafController.Del))

	safeController := controller.SafeController{}
	mux.HandleFunc("/client/safe/save", middleware.Auth(safeController.Add))
	mux.HandleFunc("/client/safe/list", middleware.Auth(safeController.List))
	mux.HandleFunc("/client/safe/remove", middleware.Auth(safeController.Del))
	mux.HandleFunc("/client/safe/query", middleware.Auth(safeController.Query))

	reverseController := controller.ReverseController{}
	mux.HandleFunc("/client/reverse/save", middleware.Auth(reverseController.Add))
	mux.HandleFunc("/client/reverse/list", middleware.Auth(reverseController.List))
	mux.HandleFunc("/client/reverse/remove", middleware.Auth(reverseController.Del))

	forwardController := controller.ForwardController{}
	mux.HandleFunc("/client/forward/save", middleware.Auth(forwardController.Add))
	mux.HandleFunc("/client/forward/list", middleware.Auth(forwardController.List))
	mux.HandleFunc("/client/forward/remove", middleware.Auth(forwardController.Del))

	giscusController := controller.GiscusController{}
	mux.HandleFunc("/client/giscus/token", middleware.Auth(giscusController.Token))

	// 实时流量大屏：WebSocket 推送入口 ws://ip:port/client/ws/stats?token=xxx
	// 浏览器 WebSocket 不能自定义请求头，所以 token 走 query 参数，鉴权在 controller 内部完成
	wsStatsController := controller.WsStatsController{}
	mux.HandleFunc("/client/ws/stats", wsStatsController.Stats)
	// 不想用 ws 的前端可以直接轮询汇总数据
	mux.HandleFunc("/client/ws/stats/summary", middleware.Auth(wsStatsController.Summary))

	muxWithRecovery := recoveryMiddleware(mux)
	webServer = &http.Server{Addr: ":" + strconv.Itoa(port), Handler: muxWithRecovery}
	log.Error(webServer.ListenAndServe())
}

// webServer 保存管理后台服务句柄，StopWebServer 时用于释放端口
var webServer *http.Server

// StopWebServer 关闭管理后台（服务停止/重启时必须释放端口，否则 -action stop 会卡住）
func StopWebServer() {
	if webServer != nil {
		_ = webServer.Close()
		webServer = nil
	}
}
