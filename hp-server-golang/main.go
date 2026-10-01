package main

import (
	"flag"
	"fmt"
	"hp-server-lib/config"
	"hp-server-lib/log"
	"hp-server-lib/net/acme"
	"hp-server-lib/net/http"
	"hp-server-lib/net/server"
	"hp-server-lib/service"
	"hp-server-lib/task"
	"hp-server-lib/web"
	syslog "log"
	"os"
	"path/filepath"
	"sync"
	"time"

	daemon "github.com/kardianos/service"
	"gopkg.in/yaml.v3"
)

// 全局变量
var logger daemon.Logger
var serviceAction string // 服务操作指令
var configPath string    // 配置文件路径

// program 实现 daemon.Interface 接口
type program struct {
	stopChan chan struct{} // 优雅退出信号通道
	// closers 收集所有可关闭的服务，Stop 时按顺序释放
	closers []namedCloser
	wg      sync.WaitGroup
	// stopOnce 保证 Stop 只执行一次：系统服务/信号可能重复触发 Stop，
	// 重复执行会 panic(close of closed channel)，导致 -action stop 直接失败
	stopOnce sync.Once
}

// stopWaitTimeout 等待业务 goroutine 退出的最长时长。
// 一旦某个 server 收不到退出信号，Stop 会永久卡在 WaitGroup 上，
// 系统服务就会一直停在"正在停止"，随后 start/stop/status/uninstall 全部超时失败。
const stopWaitTimeout = 5 * time.Second

// namedCloser 描述一个可关闭的资源（名字用于日志，Close 用于释放）。
type namedCloser struct {
	name  string
	close func()
}

// Start 服务启动入口（实现接口）
func (p *program) Start(s daemon.Service) error {
	if daemon.Interactive() {
		logger.Info("服务以交互模式启动")
	} else {
		logger.Info("服务启动成功")
	}
	go p.run() // 启动核心业务逻辑
	return nil
}

// Stop 服务停止入口（实现接口）
func (p *program) Stop(s daemon.Service) error {
	p.stopOnce.Do(func() {
		if daemon.Interactive() {
			logger.Info("服务以交互模式停止")
		} else {
			logger.Info("服务正在停止")
		}
		// 收尾：先按注册顺序反向关闭所有服务，再通知 run 退出
		p.shutdown()
		close(p.stopChan) // 发送退出信号
		p.waitDone()
	})
	return nil
}

// waitDone 带超时地等待业务 goroutine 退出，超时后不再阻塞，
// 避免 Stop 永久卡住使 -action stop/start/uninstall 全部失效
func (p *program) waitDone() {
	done := make(chan struct{})
	go func() {
		defer close(done)
		p.wg.Wait()
	}()
	select {
	case <-done:
		logger.Info("所有服务已正常退出")
	case <-time.After(stopWaitTimeout):
		logger.Errorf("等待服务退出超时（%v），强制结束停止流程", stopWaitTimeout)
	}
}

// shutdown 统一关闭所有 server（先关 listener，再让 goroutine 退出）
func (p *program) shutdown() {
	for i := len(p.closers) - 1; i >= 0; i-- {
		c := p.closers[i]
		if c.close == nil {
			continue
		}
		logger.Infof("正在关闭 %s ...", c.name)
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Errorf("关闭 %s panic: %v", c.name, r)
				}
			}()
			c.close()
		}()
	}
}

// register 注册一个可关闭的资源
func (p *program) register(name string, fn func()) {
	if fn == nil {
		return
	}
	p.closers = append(p.closers, namedCloser{name: name, close: fn})
}

// run 核心业务逻辑（整合原有所有服务启动逻辑）
func (p *program) run() {
	// 1. 加载配置文件
	if err := loadConfig(); err != nil {
		logger.Error(fmt.Sprintf("配置加载失败：%v", err))
		return
	}
	logger.Info("配置文件加载成功")

	// 2. 启动各类服务（带退出信号监听）
	p.starServer()

	// 阻塞等待退出信号
	<-p.stopChan
	logger.Info("核心业务逻辑已停止")
}

// loadConfig 加载并解析配置文件
func loadConfig() error {
	// 处理配置文件路径（支持相对/绝对路径）
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("配置文件路径解析失败：%v", err)
	}

	// 读取配置文件
	data, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败：%v（路径：%s）", err, absPath)
	}

	// 解析YAML配置
	if err := yaml.Unmarshal(data, &config.ConfigData); err != nil {
		return fmt.Errorf("解析YAML配置失败：%v", err)
	}
	return nil
}

// startServer 启动服务
func (p *program) starServer() {
	//指令控制
	cmdServer := server.NewCmdServer()
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		cmdServer.StartServer(config.ConfigData.Cmd.Port)
	}()
	p.register("cmd-server", func() { cmdServer.CLose() })
	//数据传输方式1
	quicServer := server.NewHpQuicServer(server.NewHPHandler())
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		quicServer.StartServer(config.ConfigData.Tunnel.Port)
	}()
	p.register("quic-server", func() { quicServer.CLose() })
	//数据传输方式2
	hpTcpServer := server.NewHPTcpServer(server.NewHPHandler())
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		hpTcpServer.StartServer(config.ConfigData.Tunnel.Port)
	}()
	p.register("hp-tcp-server", func() { hpTcpServer.CLose() })
	//管理后台
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		web.StartWebServer(config.ConfigData.Admin.Port)
	}()
	// 管理后台必须注册关闭方法：ListenAndServe 会一直阻塞，
	// 不 Close 的话 p.wg 永远等不到 Done，Stop 会卡死导致所有 -action 操作失效
	p.register("web-server", web.StopWebServer)
	//初始化正向代理服务
	go service.InitForward()
	if config.ConfigData.Tunnel.OpenDomain {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			http.StartHttpServer()
		}()
		p.register("http-proxy-server", http.StopHttpServer)
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			http.StartHttpsServer()
		}()
		p.register("https-proxy-server", http.StopHttpsServer)
		//缓存域名配置
		go service.InitDomainCache()
		go service.InitReverseECache()
		//acme挑战
		go func() {
			err2 := acme.StartAcmeServer(config.ConfigData.Acme.Email, config.ConfigData.Acme.HttpPort)
			if err2 != nil {
				logger.Error("证书申请服务启动失败..." + err2.Error())
			} else {
				task.StartSslTask()
			}
		}()
	}
	// 监听退出信号（Stop 中会通过 shutdown 统一关资源）
	<-p.stopChan
	logger.Info("服务正在关闭...")
}

// init 初始化命令行参数
func init() {
	// 定义命令行参数
	flag.StringVar(&serviceAction, "action", "", "服务操作：install/start/stop/uninstall/status")
	flag.StringVar(&configPath, "conf", "app.yml", "配置文件路径（默认：app.yml）")

	// 自定义帮助信息
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "使用方法：%s [参数]\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "参数说明：")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\n示例：")
		fmt.Fprintln(os.Stderr, "  交互模式运行：", os.Args[0], "-conf app.yml")
		fmt.Fprintln(os.Stderr, "  安装服务：", os.Args[0], "-conf app.yml -action install")
		fmt.Fprintln(os.Stderr, "  启动服务：", os.Args[0], "-action start")
		fmt.Fprintln(os.Stderr, "  停止服务：", os.Args[0], "-action stop")
		fmt.Fprintln(os.Stderr, "  查看状态：", os.Args[0], "-action status")
		fmt.Fprintln(os.Stderr, "  卸载服务：", os.Args[0], "-action uninstall")
	}
}

func main() {
	// 解析命令行参数（仅解析一次）
	flag.Parse()

	// 初始化程序实例
	prg := &program{
		stopChan: make(chan struct{}),
	}
	workDir, err := os.Getwd()
	if err != nil {
		syslog.Fatalf("获取当前工作目录失败：%v", err)
	}
	// 配置系统服务参数
	serviceConfig := &daemon.Config{
		Name:             "hp-lite-server", // 服务唯一标识
		DisplayName:      "hp-lite-server", // 服务显示名称
		Description:      "hp-lite-server 核心服务（含隧道、管理后台、证书管理等功能）",
		Arguments:        []string{"-conf", configPath}, // 固化配置文件路径
		WorkingDirectory: workDir,
	}

	// 创建服务实例
	s, err := daemon.New(prg, serviceConfig)
	if err != nil {
		syslog.Fatalf("服务创建失败：%v", err)
	}

	// 初始化日志（整合系统服务日志）
	logger, err = s.Logger(nil)
	if err != nil {
		syslog.Fatalf("日志初始化失败：%v", err)
	}
	log.Log = logger
	// 执行服务操作
	switch serviceAction {
	case "install":
		if err := s.Install(); err != nil {
			syslog.Fatalf("服务安装失败：%v", err)
		}
		syslog.Printf("✅ 服务安装成功！服务名称：%s", serviceConfig.Name)
		syslog.Printf("📌 配置文件路径：%s", configPath)
		syslog.Println("💡 后续操作：")
		syslog.Println("   启动服务：", os.Args[0], "-action start")
		syslog.Println("   停止服务：", os.Args[0], "-action stop")
		syslog.Println("   查看状态：", os.Args[0], "-action status")
		syslog.Println("   卸载服务：", os.Args[0], "-action uninstall")
		syslog.Println("   [注意事项]：")
		syslog.Println("   1、安装服务后请不要删除当前文件，如果需要删除当前文件，请先停止服务、然后在卸载服务、最后在删除文件")
		syslog.Println("   2、更新程序前请先停止服务再替换程序然后再启动服务")

	case "start":
		if err := s.Start(); err != nil {
			syslog.Fatalf("服务启动失败：%v", err)
		}
		syslog.Println("✅ 服务启动成功！可通过 -action status 查看状态")

	case "stop":
		if err := s.Stop(); err != nil {
			syslog.Fatalf("服务停止失败：%v", err)
		}
		syslog.Println("✅ 服务停止成功！")

	case "uninstall":
		if err := s.Uninstall(); err != nil {
			syslog.Fatalf("服务卸载失败：%v", err)
		}
		syslog.Println("✅ 服务卸载成功！")

	case "status":
		status, err := s.Status()
		if err != nil {
			syslog.Fatalf("获取服务状态失败：%v", err)
		}
		switch status {
		case daemon.StatusRunning:
			syslog.Println("🟢 服务状态：运行中")
		case daemon.StatusStopped:
			syslog.Println("🔴 服务状态：已停止")
		default:
			syslog.Printf("🟡 服务状态：%v", status)
		}

	case "":
		// 无操作 → 交互模式运行
		syslog.Printf("🚀 以交互模式启动（配置文件：%s）", configPath)
		if err := s.Run(); err != nil {
			syslog.Fatalf("交互模式运行失败：%v", err)
		}

	default:
		// 未知指令必须报错，否则"加了 -action 却什么都没发生"，看起来就像命令失效
		fmt.Fprintf(os.Stderr, "未知的操作指令：%s\n\n", serviceAction)
		flag.Usage()
		os.Exit(2)
	}
}
