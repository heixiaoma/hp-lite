package controller

import (
	"encoding/json"
	"hp-server-lib/bean"
	"hp-server-lib/service"
	"net/http"
)

type LoginController struct {
	service.UserService
}

func (receiver LoginController) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var msg bean.ReqLogin
	// 解析请求体中的JSON数据
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	login := receiver.Login(msg)
	if login != nil {
		json.NewEncoder(w).Encode(bean.ResOk(login))
		return
	} else {
		json.NewEncoder(w).Encode(bean.ResError("登陆失败"))
	}
}
