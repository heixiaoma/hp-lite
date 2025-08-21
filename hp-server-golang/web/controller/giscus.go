package controller

import (
	"encoding/json"
	"hp-server-lib/bean"
	"io"
	"net/http"
	"strings"
)

type GiscusController struct {
}

func (receiver GiscusController) Token(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	session := queryParams.Get("session")
	if len(session) > 0 {
		url := "https://giscus.app/api/oauth/token"
		payload := strings.NewReader("{\"session\":\"" + session + "\"}")
		req, _ := http.NewRequest("POST", url, payload)
		req.Header.Add("Accept", "*/*")
		req.Header.Add("Accept-Encoding", "gzip, deflate, br")
		req.Header.Add("User-Agent", "PostmanRuntime-ApipostRuntime/1.1.0")
		req.Header.Add("Connection", "keep-alive")
		req.Header.Add("Content-Type", "application/json")
		res, _ := http.DefaultClient.Do(req)
		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)
		json.NewEncoder(w).Encode(bean.ResOk(string(body)))
		return
	} else {
		json.NewEncoder(w).Encode(bean.ResError("失败"))
	}
}
