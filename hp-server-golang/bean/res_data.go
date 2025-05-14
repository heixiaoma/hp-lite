package bean

type ResData struct {
	Code int `json:"code"`

	Msg string `json:"msg"`

	Data interface{} `json:"data"`
}

func ResOk(data interface{}) *ResData {
	return &ResData{
		Code: 200,
		Msg:  "操作成功",
		Data: data,
	}
}

func ResError(msg string) *ResData {
	return &ResData{
		Code: -1,
		Msg:  msg,
	}
}
func ResErrorCode(code int, msg string) *ResData {
	return &ResData{
		Code: code,
		Msg:  msg,
	}
}
