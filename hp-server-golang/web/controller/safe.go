package controller

import (
	"encoding/json"
	"hp-server-lib/bean"
	"hp-server-lib/entity"
	"hp-server-lib/service"
	"hp-server-lib/util"
	"net/http"
	"strconv"
)

type SafeController struct {
	service.UserSafeService
}

func (receiver SafeController) getUserId(w http.ResponseWriter, r *http.Request) (int, error) {
	token := r.Header.Get("token")
	userId, _, _, err := util.DecodeToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(bean.ResErrorCode(-2, "用户权限校验失败"))
		return 0, err
	}
	return userId, nil
}

func (receiver SafeController) Add(w http.ResponseWriter, r *http.Request) {
	userId, err := receiver.getUserId(w, r)
	if err == nil {
		var msg entity.UserSafeEntity
		// 解析请求体中的JSON数据
		err := json.NewDecoder(r.Body).Decode(&msg)
		if err != nil {
			println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = receiver.AddData(userId, msg)
		if err != nil {
			json.NewEncoder(w).Encode(bean.ResError(err.Error()))
			return
		}
		json.NewEncoder(w).Encode(bean.ResOk(nil))
	}
}

func (receiver SafeController) List(w http.ResponseWriter, r *http.Request) {
	userId, err := receiver.getUserId(w, r)
	if err == nil {
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
		json.NewEncoder(w).Encode(bean.ResOk(receiver.ListData(userId, pageInt, pageSizeInt)))
	}
}

func (receiver SafeController) Del(w http.ResponseWriter, r *http.Request) {
	_, err := receiver.getUserId(w, r)
	if err == nil {
		queryParams := r.URL.Query()
		id := queryParams.Get("id")
		idInt, _ := strconv.Atoi(id)
		receiver.RemoveData(idInt)
		json.NewEncoder(w).Encode(bean.ResOk(nil))
	}
}

func (receiver SafeController) Query(w http.ResponseWriter, r *http.Request) {
	userId, err := receiver.getUserId(w, r)
	if err == nil {
		queryParams := r.URL.Query()
		keyword := queryParams.Get("keyword")
		json.NewEncoder(w).Encode(bean.ResOk(receiver.SafeListByKey(userId, keyword)))
	}
}
