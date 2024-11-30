package util

import (
	"bytes"
	"github.com/olekukonko/tablewriter"
	"log"
)

func Print(msg string) {
	log.Println(msg)
}

func PrintStatus(data [][]string) string {
	if len(data) == 0 {
		return "暂无穿配置"
	}
	// 创建表格
	buffer := bytes.NewBuffer(nil)
	table := tablewriter.NewWriter(buffer)
	// 设置标题行
	table.SetHeader([]string{"描述", "内容"})
	for _, wear := range data {
		if wear == nil {
			return "暂无穿配置"
		}
		msg := []string{"", ""}
		msg[0] = wear[0]
		msg[1] = wear[1]
		table.Append(msg)
	}
	// 渲染表格
	table.Render()
	result := buffer.String()
	return "\r\n" + result
}
