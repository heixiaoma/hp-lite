package controller

import (
	"embed"
	"net/http"
	"strings"
)

type StaticController struct {
	Content embed.FS
}

func (s StaticController) Static(w http.ResponseWriter, r *http.Request) {
	// 确保路径去除前缀（这里的"/"是从根路径开始的）
	filePath := r.URL.Path
	if filePath == "/" {
		filePath = "/index.html" // 默认返回 index.html
	}
	// 从嵌入文件系统读取文件
	data, err := s.Content.ReadFile("static" + filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// 设置正确的 Content-Type（根据文件扩展名）
	switch {
	case strings.HasSuffix(filePath, ".html"):
		w.Header().Set("Content-Type", "text/html")
	case strings.HasSuffix(filePath, ".css"):
		w.Header().Set("Content-Type", "text/css")
	case strings.HasSuffix(filePath, ".js"):
		w.Header().Set("Content-Type", "application/javascript")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	w.Write(data)
}
