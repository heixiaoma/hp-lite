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

type DomainController struct {
	service.DomainService
}

func (receiver DomainController) getUserId(w http.ResponseWriter, r *http.Request) (int, error) {
	token := r.Header.Get("token")
	userId, _, _, err := util.DecodeToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(bean.ResErrorCode(-2, "用户权限校验失败"))
		return 0, err
	}
	return userId, nil
}

func (receiver DomainController) GetDomainList(w http.ResponseWriter, r *http.Request) {
	id, err := receiver.getUserId(w, r)
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
		json.NewEncoder(w).Encode(bean.ResOk(receiver.DomainList(id, pageInt, pageSizeInt)))
	}
}

func (receiver DomainController) RemoveDomain(w http.ResponseWriter, r *http.Request) {
	_, err := receiver.getUserId(w, r)
	if err == nil {
		queryParams := r.URL.Query()
		id := queryParams.Get("id")
		idInt, _ := strconv.Atoi(id)
		json.NewEncoder(w).Encode(bean.ResOk(receiver.RemoveData(idInt)))
	}
}

func (receiver DomainController) Query(w http.ResponseWriter, r *http.Request) {
	userId, err := receiver.getUserId(w, r)
	if err == nil {
		queryParams := r.URL.Query()
		keyword := queryParams.Get("keyword")
		json.NewEncoder(w).Encode(bean.ResOk(receiver.DomainListByKey(userId, keyword)))
	}
}

func (receiver DomainController) Gen(w http.ResponseWriter, r *http.Request) {
	_, err := receiver.getUserId(w, r)
	if err == nil {
		queryParams := r.URL.Query()
		id := queryParams.Get("id")
		idInt, _ := strconv.Atoi(id)
		json.NewEncoder(w).Encode(bean.ResOk(receiver.GenSsl(idInt)))
	}
}

func (receiver DomainController) Add(w http.ResponseWriter, r *http.Request) {
	userId, err := receiver.getUserId(w, r)
	if err == nil {
		var msg entity.UserDomainEntity
		// 解析请求体中的JSON数据
		err := json.NewDecoder(r.Body).Decode(&msg)
		if err != nil {
			println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		msg.UserId = &userId
		err = receiver.AddData(msg)
		if err == nil {
			json.NewEncoder(w).Encode(bean.ResOk(nil))
			return
		}
		json.NewEncoder(w).Encode(bean.ResError(err.Error()))
	}
}
