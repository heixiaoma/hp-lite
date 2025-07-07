package controller

import (
	"encoding/json"
	"hp-server-lib/bean"
	"hp-server-lib/entity"
	"hp-server-lib/service"
	"hp-server-lib/util"
	"net/http"
	"strconv"
	"strings"
)

type WafController struct {
	service.UserWafService
}

func (receiver WafController) checkUser(w http.ResponseWriter, r *http.Request) bool {
	token := r.Header.Get("token")
	_, s, _, err := util.DecodeToken(token)
	if err == nil && strings.Compare(s, "ADMIN") == 0 {
		return true
	}
	json.NewEncoder(w).Encode(bean.ResErrorCode(-2, "用户权限校验失败"))
	return false
}

func (receiver WafController) Add(w http.ResponseWriter, r *http.Request) {
	if receiver.checkUser(w, r) {
		var msg entity.UserWafEntity
		// 解析请求体中的JSON数据
		err := json.NewDecoder(r.Body).Decode(&msg)
		if err != nil {
			println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = receiver.AddData(msg)
		if err != nil {
			json.NewEncoder(w).Encode(bean.ResError(err.Error()))
			return
		}
		json.NewEncoder(w).Encode(bean.ResOk(nil))
	}
}

func (receiver WafController) List(w http.ResponseWriter, r *http.Request) {
	if receiver.checkUser(w, r) {
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
}

func (receiver WafController) Del(w http.ResponseWriter, r *http.Request) {
	if receiver.checkUser(w, r) {
		queryParams := r.URL.Query()
		id := queryParams.Get("id")
		idInt, _ := strconv.Atoi(id)
		receiver.RemoveData(idInt)
		json.NewEncoder(w).Encode(bean.ResOk(nil))
	}
}
