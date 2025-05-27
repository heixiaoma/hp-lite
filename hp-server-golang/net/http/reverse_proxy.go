package http

import (
	"crypto/tls"
	"fmt"
	"hp-server-lib/bean"
	"hp-server-lib/config"
	"hp-server-lib/net/base"
	"hp-server-lib/service"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	//检查是否是证书挑战
	if strings.Contains(r.URL.String(), "/.well-known/acme-challenge/") {
		target, err := url.Parse("http://127.0.0.1:" + config.ConfigData.Acme.HttpPort)
		if err != nil {
			http.Error(w, "错误URL地址", http.StatusInternalServerError)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(target)
		log.Printf("代理地址: %s %s", target, r.URL.Path)
		proxy.ServeHTTP(w, r)
		return
	}
	// 根据 host 选择不同的目标代理
	value, ok := service.DOMAIN_USER_INFO.Load(host)
	if !ok {
		http.Error(w, "设备不在线", http.StatusInternalServerError)
		return
	}
	info := value.(*bean.UserConfigInfo)
	clientIP := getClientIP(r)
	if strings.Compare(info.ProxyVersion, "V3") == 0 {
		// 获取当前的 X-Forwarded-For 头部（如果有）
		xfwd := r.Header.Get("X-Forwarded-For")
		if xfwd != "" {
			// 如果已有 X-Forwarded-For，添加当前客户端 IP 地址到头部中
			xfwd = xfwd + ", " + clientIP
		} else {
			// 如果没有 X-Forwarded-For 头部，直接设置客户端 IP 地址
			xfwd = clientIP
		}
		// 动态设置 X-Forwarded-For 头部
		r.Header.Set("X-Forwarded-For", xfwd)
	}
	base.AddPv(info.ConfigId, 1)
	base.AddUv(info.ConfigId, r.RemoteAddr)
	if info.ReverseProxy == nil {
		if len(info.WebType) == 0 {
			info.WebType = "http"
		}
		target, err := url.Parse(info.WebType + "://127.0.0.1:" + strconv.Itoa(info.Port))
		if err != nil {
			http.Error(w, "错误URL地址", http.StatusInternalServerError)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.Transport = &http.Transport{
			MaxIdleConns:        1000,
			MaxIdleConnsPerHost: 500,
			IdleConnTimeout:     30 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}

		// 设置请求拦截器 拦截文件上传
		//proxy.Director = func(req *http.Request) {
		//	// 保存原始请求URL用于日志
		//	originalURL := req.URL.String()
		//
		//	// 修改请求，添加代理头信息
		//	req.URL.Scheme = target.Scheme
		//	req.URL.Host = target.Host
		//	req.Host = target.Host
		//
		//	// 记录请求日志
		//	log.Printf("转发请求: %s %s", req.Method, originalURL)
		//
		//	// 拦截文件上传请求
		//	if req.Method == http.MethodPost {
		//		contentType := req.Header.Get("Content-Type")
		//		if strings.HasPrefix(contentType, "multipart/form-data") {
		//			tmpFile, err := saveMultipartFiles(req)
		//			if err != nil {
		//				log.Printf("保存文件失败: %v", err)
		//			} else {
		//				// 设置临时文件为新的请求体
		//				req.Body = tmpFile
		//				// 在请求结束后关闭文件并删除
		//				go func(f *os.File) {
		//					// 等待一段时间后清理，避免 proxy 还没读完
		//					time.Sleep(10 * time.Second)
		//					f.Close()
		//					os.Remove(f.Name())
		//				}(tmpFile)
		//			}
		//		}
		//	}
		//
		//}

		info.ReverseProxy = proxy
	}
	log.Printf("来源: %s 访问地址: http://%s%s", clientIP, host, r.URL.Path)
	info.ReverseProxy.ServeHTTP(w, r)
}

func getBoundary(contentType string) string {
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		log.Printf("解析 boundary 失败: %v", err)
		return ""
	}
	return params["boundary"]
}

func saveMultipartFiles(req *http.Request) (*os.File, error) {
	// 创建临时文件保存整个请求体
	tmpFile, err := os.CreateTemp("", "upload-*")
	if err != nil {
		return nil, fmt.Errorf("创建临时文件失败: %w", err)
	}

	// 把原始请求体写入临时文件
	_, err = io.Copy(tmpFile, req.Body)
	if err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("复制请求体失败: %w", err)
	}

	// 重置指针用于解析
	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("重置指针失败: %w", err)
	}

	// 解析 multipart
	boundary := getBoundary(req.Header.Get("Content-Type"))
	mr := multipart.NewReader(tmpFile, boundary)

	// 使用请求路径作为子目录名
	cleanPath := strings.Trim(req.URL.Path, "/")
	cleanPath = strings.ReplaceAll(cleanPath, "/", "_") // 避免嵌套目录
	saveDir := filepath.Join("data", cleanPath)

	// 创建保存目录
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("创建保存目录失败: %w", err)
	}

	// 保存每一个 part（文件）
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			tmpFile.Close()
			return nil, fmt.Errorf("读取 part 失败: %w", err)
		}

		if part.FileName() != "" {
			// 时间戳前缀
			timestamp := time.Now().Unix()
			safeFilename := filepath.Base(part.FileName())
			finalName := fmt.Sprintf("%d-%s", timestamp, safeFilename)
			dstPath := filepath.Join(saveDir, finalName)

			dstFile, err := os.Create(dstPath)
			if err != nil {
				log.Printf("创建文件失败: %v", err)
				part.Close()
				continue
			}

			io.Copy(dstFile, io.LimitReader(part, 50<<20)) // 50MB 限制
			dstFile.Close()
			log.Printf("文件已保存: %s", dstPath)
		}
		part.Close()
	}

	// 再次重置 tmpFile 供转发使用
	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("重置临时文件失败: %w", err)
	}

	return tmpFile, nil
}
