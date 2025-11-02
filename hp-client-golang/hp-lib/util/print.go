package util

import (
	"bytes"
	"hp-lib/bean"
	"log"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
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
	symbols := tw.NewSymbolCustom("Nature").
		WithRow("-").
		WithColumn("|").
		WithTopLeft("🌱").
		WithTopMid("🌿").
		WithTopRight("🌱").
		WithMidLeft("🍃").
		WithCenter("❀").
		WithMidRight("🍃").
		WithBottomLeft("🌻").
		WithBottomMid("🌾").
		WithBottomRight("🌻")

	table := tablewriter.NewTable(buffer, tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{Symbols: symbols})))
	// 设置标题行
	table.Header([]string{"远端服务", "内网服务", "隧道类型", "状态"})

	for _, wear := range data {
		if wear == nil {
			return "暂无穿配置"
		}
		msg := []string{"", "", "", ""}
		msg[0] = wear.ServerIp + ":" + strconv.Itoa(wear.RemotePort)
		msg[1] = wear.LocalAddress
		msg[2] = wear.TunType
		msg[3] = strconv.FormatBool(wear.Status)
		table.Append(msg)
	}
	// 渲染表格
	table.Render()
	result := buffer.String()
	return "\r\n" + result
}
