package util

import (
	"bytes"
	"github.com/olekukonko/tablewriter"
	"hp-lib/bean"
	"log"
	"strconv"
)

func Print(msg string) {
	log.Println(msg)
}

func PrintStatus(data []*bean.LocalInnerWear) string {
	if len(data) == 0 {
		return "暂无穿配置"
	}
	// 创建表格
	buffer := bytes.NewBuffer(nil)
	table := tablewriter.NewWriter(buffer)
	// 设置标题行
	table.SetHeader([]string{"远端服务", "内网服务", "映射类型", "状态"})
	for _, wear := range data {
		if wear == nil {
			return "暂无穿配置"
		}
		msg := []string{"", "", "", ""}
		msg[0] = wear.ServerIp
		msg[1] = wear.LocalIp + ":" + strconv.Itoa(wear.LocalPort)
		switch wear.ConnectType {
		case bean.TCP:
			msg[2] = "TCP"
		case bean.UDP:
			msg[2] = "UDP"
		case bean.TCP_UDP:
			msg[2] = "TCP_UDP"
		}
		msg[3] = strconv.FormatBool(wear.Status)
		table.Append(msg)
	}
	// 渲染表格
	table.Render()
	result := buffer.String()
	return "\r\n" + result
}
