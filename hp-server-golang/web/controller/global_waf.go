package controller

import (
	"encoding/json"
	"hp-server-lib/bean"
	"hp-server-lib/entity"
	"hp-server-lib/service"
	"net/http"
	"strconv"
)

/* 全局 Web 安全规则。
   和 SafeController 的区别：那边是用户自己的规则集，要绑到穿透 / 反向代理上才生效；
   这里管的是系统级规则，对所有进站流量生效。
   所以路由一律走 middleware.Admin —— 只有管理员能增删改（见 web_server.go）。 */

type GlobalWafController struct {
	service.GlobalWafService
}

func (receiver GlobalWafController) Add(w http.ResponseWriter, r *http.Request) {
	var msg entity.GlobalWafEntity
	// 解析请求体中的JSON数据
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// 规则编译不过会带错误回来：直接透传给前端，管理员好对着改
	err = receiver.AddData(msg)
	if err != nil {
		json.NewEncoder(w).Encode(bean.ResError(err.Error()))
		return
	}
	json.NewEncoder(w).Encode(bean.ResOk(nil))
}

func (receiver GlobalWafController) List(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	page := queryParams.Get("current")
	pageSize := queryParams.Get("pageSize")
	pageInt, _ := strconv.Atoi(page)
	pageSizeInt, _ := strconv.Atoi(pageSize)
	if pageInt == 0 {
		pageInt = 1
	}
	if pageSizeInt == 0 {
		pageSizeInt = 10
	}
	json.NewEncoder(w).Encode(bean.ResOk(receiver.ListData(pageInt, pageSizeInt)))
}

func (receiver GlobalWafController) Del(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	id := queryParams.Get("id")
	idInt, _ := strconv.Atoi(id)
	receiver.RemoveData(idInt)
	json.NewEncoder(w).Encode(bean.ResOk(nil))
}
