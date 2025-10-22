package http

import (
	"bytes"
	"html/template"
	"os"
	"sync"
)

var (
	cachedTemplate     string
	cachedTemplateOnce sync.Once
)

var (
	wafTpl     *template.Template
	wafTplOnce sync.Once
)

func DeviceNotFound() string {
	// 用 sync.Once 保证只加载一次模板
	cachedTemplateOnce.Do(func() {
		path := "./template/not_found.html"
		data, err := os.ReadFile(path)
		if err != nil {
			// 文件不存在，使用默认模板
			cachedTemplate = defaultDeviceNotFound()
			return
		}
		cachedTemplate = string(data)
	})
	return cachedTemplate
}

func defaultDeviceNotFound() string {
	return `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>设备不在线</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
        }
        body {
            min-height: 100vh;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            background-color: #f8fafc;
            color: #334155;
            padding: 2rem;
        }
        .error-container {
            max-width: 500px;
            width: 100%;
            text-align: center;
            padding: 2rem;
            background-color: #ffffff;
            border-radius: 12px;
            box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
        }
        .error-icon {
            width: 120px;
            height: 120px;
            margin: 0 auto 2rem;
            position: relative;
        }
        .error-icon::before {
            content: "";
            position: absolute;
            width: 100%;
            height: 100%;
            background-color: #fee2e2;
            border-radius: 50%;
            z-index: 1;
        }
        .error-icon::after {
            content: "📶";
            position: absolute;
            font-size: 50px;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            z-index: 2;
            opacity: 0.8;
        }
        h1 {
            font-size: 1.8rem;
            color: #dc2626;
            margin-bottom: 1rem;
            font-weight: 600;
        }
        .error-message {
            color: #64748b;
            margin-bottom: 2rem;
            line-height: 1.6;
            font-size: 1rem;
        }
        .details {
            background-color: #f8fafc;
            border-radius: 8px;
            padding: 1rem;
            margin-bottom: 2rem;
            text-align: left;
            font-size: 0.9rem;
            color: #475569;
        }
        .detail-item {
            margin-bottom: 0.5rem;
            display: flex;
            align-items: center;
        }
        .detail-item::before {
            content: "•";
            color: #94a3b8;
            margin-right: 0.5rem;
        }
        .action-buttons {
            display: flex;
            flex-direction: column;
            gap: 1rem;
        }
        button {
            padding: 0.8rem 1.5rem;
            border: none;
            border-radius: 8px;
            font-size: 1rem;
            font-weight: 500;
            cursor: pointer;
            transition: all 0.2s ease;
        }
        .primary-btn {
            background-color: #2563eb;
            color: white;
        }
        .primary-btn:hover {
            background-color: #1d4ed8;
            box-shadow: 0 4px 12px rgba(37, 99, 235, 0.25);
        }
        .secondary-btn {
            background-color: #f1f5f9;
            color: #334155;
        }
        .secondary-btn:hover {
            background-color: #e2e8f0;
        }
        .footer {
            margin-top: 2rem;
            color: #94a3b8;
            font-size: 0.85rem;
        }
        @media (max-width: 480px) {
            .error-container {
                padding: 1.5rem;
            }
            .error-icon {
                width: 100px;
                height: 100px;
            }
            h1 {
                font-size: 1.5rem;
            }
        }
    </style>
</head>
<body>
    <div class="error-container">
        <div class="error-icon"></div>
        <h1>设备不在线</h1>
        <p class="error-message">
            无法连接到设备，请检查网络连接或设备状态后重试。
        </p>
        <div class="details">
            <div class="detail-item">检查设备是否已开机并正常运行</div>
            <div class="detail-item">确认网络连接是否稳定</div>
            <div class="detail-item">验证设备是否在信号覆盖范围内</div>
        </div>
        <div class="action-buttons">
            <button class="primary-btn" onclick="location.reload()">刷新页面</button>
        </div>
    </div>
    <div class="footer">
        最后检查时间: <span id="check-time"></span>
    </div>
    <script>
        // 显示最后检查时间
        const updateCheckTime = () => {
            const now = new Date();
            const timeString = now.toLocaleString('zh-CN', {
                year: 'numeric',
                month: '2-digit',
                day: '2-digit',
                hour: '2-digit',
                minute: '2-digit',
                second: '2-digit'
            });
            document.getElementById('check-time').textContent = timeString;
        };
        // 初始化时间显示
        updateCheckTime();
        // 每30秒更新一次检查时间
        setInterval(updateCheckTime, 30000);
    </script>
</body>
</html>
`
}

func Waf(str, ip string) string {
	// 模板加载一次后缓存
	wafTplOnce.Do(func() {
		path := "./template/waf.html"
		data, err := os.ReadFile(path)
		if err != nil {
			// 文件不存在，使用默认模板
			wafTpl = template.Must(template.New("default_waf").Parse(defaultWafHTML()))
			return
		}
		// 编译模板
		wafTpl = template.Must(template.New("custom_waf").Parse(string(data)))
	})

	// 渲染模板
	var buf bytes.Buffer
	_ = wafTpl.Execute(&buf, map[string]string{
		"Str": str,
		"IP":  ip,
	})
	return buf.String()
}

func defaultWafHTML() string {
	return `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>访问受限</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
        }

        body {
            min-height: 100vh;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            background-color: #f8fafc;
            color: #334155;
            padding: 2rem;
        }

        .error-container {
            max-width: 500px;
            width: 100%;
            text-align: center;
            padding: 2rem;
            background-color: #ffffff;
            border-radius: 12px;
            box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
        }

        .error-icon {
            width: 120px;
            height: 120px;
            margin: 0 auto 2rem;
            position: relative;
        }

        .error-icon::before {
            content: "";
            position: absolute;
            width: 100%;
            height: 100%;
            background-color: #fee2e2;
            border-radius: 50%;
            z-index: 1;
        }

        .error-icon::after {
            content: "🚫";
            position: absolute;
            font-size: 50px;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            z-index: 2;
            opacity: 0.8;
        }

        h1 {
            font-size: 1.8rem;
            color: #dc2626;
            margin-bottom: 1rem;
            font-weight: 600;
        }

        .error-message {
            color: #64748b;
            margin-bottom: 2rem;
            line-height: 1.6;
            font-size: 1rem;
        }

        .details {
            background-color: #f8fafc;
            border-radius: 8px;
            padding: 1rem;
            margin-bottom: 2rem;
            text-align: left;
            font-size: 0.9rem;
            color: #475569;
        }

        .detail-item {
            margin-bottom: 0.5rem;
            display: flex;
            align-items: center;
        }

        .detail-item::before {
            content: "•";
            color: #94a3b8;
            margin-right: 0.5rem;
        }

        .blocked-ip {
            font-weight: 600;
            color: #dc2626;
            background-color: #fee2e2;
            padding: 0.2rem 0.5rem;
            border-radius: 4px;
        }

        .action-buttons {
            display: flex;
            flex-direction: column;
            gap: 1rem;
        }

        button {
            padding: 0.8rem 1.5rem;
            border: none;
            border-radius: 8px;
            font-size: 1rem;
            font-weight: 500;
            cursor: pointer;
            transition: all 0.2s ease;
        }

        .primary-btn {
            background-color: #2563eb;
            color: white;
        }

        .primary-btn:hover {
            background-color: #1d4ed8;
            box-shadow: 0 4px 12px rgba(37, 99, 235, 0.25);
        }

        .secondary-btn {
            background-color: #f1f5f9;
            color: #334155;
        }

        .secondary-btn:hover {
            background-color: #e2e8f0;
        }

        .footer {
            margin-top: 2rem;
            color: #94a3b8;
            font-size: 0.85rem;
        }

        @media (max-width: 480px) {
            .error-container {
                padding: 1.5rem;
            }

            .error-icon {
                width: 100px;
                height: 100px;
            }

            h1 {
                font-size: 1.5rem;
            }
        }
    </style>
</head>
<body>
    <div class="error-container">
        <div class="error-icon"></div>
        <h1>访问受限</h1>
        <p class="error-message">
            {{.Str}}
        </p>
        <div class="details">
            <div class="detail-item">您的IP地址: <span class="blocked-ip" id="blocked-ip">{{.IP}}</span></div>
            <div class="detail-item">此IP因安全原因被临时或永久阻止</div>
            <div class="detail-item">如有疑问，请联系系统管理员</div>
        </div>
        
        <div class="action-buttons">
            <button class="primary-btn" onclick="location.reload()">刷新页面</button>
        </div>
    </div>
    
    <div class="footer">
        访问被阻止时间: <span id="block-time"></span>
    </div>

    <script>
        // 显示阻止时间
        const updateBlockTime = () => {
            const now = new Date();
            const timeString = now.toLocaleString('zh-CN', {
                year: 'numeric',
                month: '2-digit',
                day: '2-digit',
                hour: '2-digit',
                minute: '2-digit',
                second: '2-digit'
            });
            document.getElementById('block-time').textContent = timeString;
        };
        // 初始化阻止时间
        updateBlockTime();
    </script>
</body>
</html>
`

}
