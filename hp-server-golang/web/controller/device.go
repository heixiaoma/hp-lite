package controller

import (
	"encoding/json"
	"hp-server-lib/bean"
	"hp-server-lib/service"
	"hp-server-lib/util"
	"net/http"
)

type DeviceController struct {
	service.DeviceService
}

func (receiver DeviceController) getUserId(w http.ResponseWriter, r *http.Request) (int, error) {
	token := r.Header.Get("token")
	userId, _, _, err := util.DecodeToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(bean.ResErrorCode(-2, "用户权限校验失败"))
		return 0, err
	}
	return userId, nil
}
func (receiver DeviceController) Add(w http.ResponseWriter, r *http.Request) {
	userId, err := receiver.getUserId(w, r)
	if err == nil {
		var msg bean.ReqDeviceInfo
		// 解析请求体中的JSON数据
		err := json.NewDecoder(r.Body).Decode(&msg)
		if err != nil {
			println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = receiver.AddData(userId, msg)
		if err == nil {
			json.NewEncoder(w).Encode(bean.ResOk(nil))
			return
		}
		json.NewEncoder(w).Encode(bean.ResError(err.Error()))
	}
}

func (receiver DeviceController) Update(w http.ResponseWriter, r *http.Request) {
	_, err := receiver.getUserId(w, r)
	if err == nil {
		var msg bean.ReqDeviceInfo
		// 解析请求体中的JSON数据
		err := json.NewDecoder(r.Body).Decode(&msg)
		if err != nil {
			println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = receiver.UpdateData(msg)
		if err == nil {
			json.NewEncoder(w).Encode(bean.ResOk(nil))
			return
		}
		json.NewEncoder(w).Encode(bean.ResError(err.Error()))
	}
}

func (receiver DeviceController) List(w http.ResponseWriter, r *http.Request) {
	userId, err := receiver.getUserId(w, r)
	if err == nil {
		json.NewEncoder(w).Encode(bean.ResOk(receiver.ListData(userId)))
	}
}

func (receiver DeviceController) Del(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	deviceId := queryParams.Get("deviceId")
	err := receiver.RemoveData(deviceId)
	if err == nil {
		json.NewEncoder(w).Encode(bean.ResOk(nil))
		return
	}
	json.NewEncoder(w).Encode(bean.ResError(err.Error()))
}

func (receiver DeviceController) Stop(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	deviceId := queryParams.Get("deviceId")
	json.NewEncoder(w).Encode(bean.ResOk(receiver.StopData(deviceId)))
}
