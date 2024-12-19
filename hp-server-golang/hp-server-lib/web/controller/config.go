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

type ConfigController struct {
	service.ConfigService
}

func (receiver ConfigController) getUserId(w http.ResponseWriter, r *http.Request) (int, error) {
	token := r.Header.Get("token")
	userId, _, _, err := util.DecodeToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(bean.ResErrorCode(-2, "用户权限校验失败"))
		return 0, err
	}
	return userId, nil
}

func (receiver ConfigController) GetDeviceKey(w http.ResponseWriter, r *http.Request) {
	id, err := receiver.getUserId(w, r)
	if err == nil {
		json.NewEncoder(w).Encode(bean.ResOk(receiver.DeviceKey(id)))
	}
}

func (receiver ConfigController) GetConfigList(w http.ResponseWriter, r *http.Request) {
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
		json.NewEncoder(w).Encode(bean.ResOk(receiver.ConfigList(id, pageInt, pageSizeInt)))
	}
}

func (receiver ConfigController) RemoveConfig(w http.ResponseWriter, r *http.Request) {
	_, err := receiver.getUserId(w, r)
	if err == nil {
		queryParams := r.URL.Query()
		configId := queryParams.Get("configId")
		configIdInt, _ := strconv.Atoi(configId)
		json.NewEncoder(w).Encode(bean.ResOk(receiver.RemoveData(configIdInt)))
	}
}

func (receiver ConfigController) Add(w http.ResponseWriter, r *http.Request) {
	_, err := receiver.getUserId(w, r)
	if err == nil {
		var msg entity.UserConfigEntity
		// 解析请求体中的JSON数据
		err := json.NewDecoder(r.Body).Decode(&msg)
		if err != nil {
			println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = receiver.AddData(msg)
		if err == nil {
			json.NewEncoder(w).Encode(bean.ResOk(nil))
			return
		}
		json.NewEncoder(w).Encode(bean.ResError(err.Error()))
	}
}

func (receiver ConfigController) RefConfig(w http.ResponseWriter, r *http.Request) {
	_, err := receiver.getUserId(w, r)
	if err == nil {
		queryParams := r.URL.Query()
		configId := queryParams.Get("configId")
		configIdInt, _ := strconv.Atoi(configId)
		err = receiver.RefData(configIdInt)
		if err == nil {
			json.NewEncoder(w).Encode(bean.ResOk(nil))
			return
		}
		json.NewEncoder(w).Encode(bean.ResError(err.Error()))
	}
}
