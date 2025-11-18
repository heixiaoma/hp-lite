package util

import (
	"bytes"
	"hp-server-lib/log"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

func Print(msg string) {
	log.Info(msg)
}

func PrintStatus(data [][]string) string {
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
	table.Header([]string{"描述", "内容"})
	table.Bulk(data)
	table.Render()
	result := buffer.String()
	return "\r\n" + result
}
