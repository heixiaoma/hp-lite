package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	// 设置要转发的地址
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		fmt.Println("Host:", host)
		// 根据 host 选择不同的目标代理
		target, err := url.Parse("http://192.168.100.246:5666") // 可以根据 host 选择不同的后端地址
		if err != nil {
			http.Error(w, "Error parsing target URL", http.StatusInternalServerError)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.ServeHTTP(w, r)
	})
	http.ListenAndServe(":443", nil)

}
