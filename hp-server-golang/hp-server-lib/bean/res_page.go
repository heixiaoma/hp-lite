package bean

type ResPage struct {
	Total   int64       `json:"total"`
	Records interface{} `json:"records"`
}

func PageOk(Total int64, Records interface{}) *ResPage {
	return &ResPage{
		Total:   Total,
		Records: Records,
	}
}
