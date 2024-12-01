package http

import (
	"fmt"
	"hp-server-lib/bean"
	"hp-server-lib/service"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
)

func StartHttpServer() {

	mux := http.NewServeMux()

	// 使用反向代理处理所有请求
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		fmt.Println("Host:", host)
		// 根据 host 选择不同的目标代理
		value, ok := service.DOMAIN_USER_INFO.Load(host)
		if !ok {
			http.Error(w, "错误代理", http.StatusInternalServerError)
			return
		}

		info := value.(*bean.UserConfigInfo)

		target, err := url.Parse("http://127.0.0.1:" + strconv.Itoa(info.Port))
		if err != nil {
			http.Error(w, "错误URL地址", http.StatusInternalServerError)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(target)
		log.Printf("Proxying request to target: %s %s", target, r.URL.Path)
		proxy.ServeHTTP(w, r)
	})
	err := http.ListenAndServe(":80", mux)
	if err != nil {
		return
	}
}
