package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"hp-lib/net/cmd"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func init() {
	//设置中文字体:解决中文乱码问题
	fontPaths := findfont.List()
	for _, path := range fontPaths {
		log.Println(path)
		if strings.Contains(path, "Arial Unicode.ttf") || strings.Contains(path, "msyh.ttf") || strings.Contains(path, "simhei.ttf") || strings.Contains(path, "simsun.ttc") || strings.Contains(path, "simkai.ttf") {
			os.Setenv("FYNE_FONT", path)
			break
		}
	}
}

func CountLines(s string) int {
	lines := strings.Count(s, "\n")
	return lines
}
func main() {

	serverIp := ""
	serverPort := 0

	a := app.New()
	w := a.NewWindow("HP-LITE映射工具")

	serverInput := widget.NewEntry()
	serverInput.SetPlaceHolder("请输入服务地址:如 xxx.com:6666")

	deviceIdInput := widget.NewEntry()
	deviceIdInput.SetPlaceHolder("请输入设备ID...")

	var connectBtn *widget.Button
	var cmdClient *cmd.CmdClient
	var logs []string
	var list *widget.List
	list = widget.NewList(
		func() int {
			return len(logs)
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			return label
		},
		func(i widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			s := logs[i]
			lineCount := CountLines(s)
			if lineCount > 1 {
				list.SetItemHeight(i, float32(lineCount*24))
			}
			label.SetText(s)
		},
	)

	connectBtn = widget.NewButton("连接云端", func() {
		if cmdClient != nil {
			cmdClient.Close()
			cmdClient = nil
			connectBtn.SetText("连接云端")
			deviceIdInput.Enable()
			serverInput.Enable()
		} else {
			cmdClient = cmd.NewCmdClient(func(message string) {
				if len(logs) > 20 {
					logs = logs[20:]
				}
				logs = append(logs, strings.TrimSpace(message))
				list.Refresh()
				list.ScrollToBottom()
				log.Printf(message)
			})
			split := strings.Split(serverInput.Text, ":")
			serverPort, _ := strconv.Atoi(split[1])
			serverIp := split[0]
			cmdClient.Connect(serverIp, serverPort, deviceIdInput.Text)
			connectBtn.SetText("断开连接")
			deviceIdInput.Disable()
			serverInput.Disable()
		}
	})

	go func() {
		for {
			if cmdClient != nil && !cmdClient.GetStatus() && strings.Contains(connectBtn.Text, "断开连接") {
				cmdClient.Connect(serverIp, serverPort, deviceIdInput.Text)
			}
			time.Sleep(time.Duration(10) * time.Second)
		}
	}()
	vBox := container.NewVBox(
		serverInput,
		deviceIdInput,
		connectBtn,
	)
	stack := container.NewPadded(
		list,
	)
	border := container.NewBorder(vBox, nil, nil, nil, stack)
	w.SetContent(border)
	w.Resize(fyne.NewSize(700, 700))
	w.ShowAndRun()
}
