package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	photosorter "photo-sorter/internal/app/photosorter"
	"photo-sorter/internal/pkg/config"
	"photo-sorter/internal/pkg/logger"
	"photo-sorter/internal/pkg/metadata"
	"photo-sorter/internal/pkg/version"
)

var (
	srcDir     string
	dstDir     string
	workers    int
	configPath string
	showVer    bool
	cpuProfile string // CPU profile 檔案路徑
	memProfile string // 記憶體 profile 檔案路徑
)

func init() {
	flag.StringVar(&srcDir, "src", ".", "Source photo directory")
	flag.StringVar(&dstDir, "dst", ".", "Destination directory for sorted photos")
	flag.IntVar(&workers, "workers", runtime.NumCPU(), "Maximum number of concurrent workers")
	flag.StringVar(&configPath, "c", "config.yaml", "Configuration file path")
	flag.BoolVar(&showVer, "version", false, "Show version information")
	flag.StringVar(&cpuProfile, "cpuprofile", "", "CPU profile file path")
	flag.StringVar(&memProfile, "memprofile", "", "Memory profile file path")
}

func main() {
	// 解析命令列參數
	flag.Parse()

	// 顯示版本資訊
	if showVer {
		fmt.Println(version.GetVersion())
		return
	}

	// 啟動 CPU profiling
	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			log.Fatalf("Failed to create CPU profile: %v", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatalf("Failed to start CPU profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}

	// 啟動記憶體 profiling
	if memProfile != "" {
		f, err := os.Create(memProfile)
		if err != nil {
			log.Fatalf("Failed to create memory profile: %v", err)
		}
		defer f.Close()
		defer pprof.WriteHeapProfile(f)
	}

	// 載入配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 套用命令列參數
	cfg.ApplyFlags(srcDir, dstDir, workers)

	// 建立日誌記錄器
	logger, err := logger.NewLogger(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// 設定 metadata 的警告輸出
	metadata.Warnf = func(format string, args ...any) {
		logger.LogWarn(fmt.Sprintf(format, args...))
	}

	// 建立應用程式
	app := photosorter.NewApp(cfg, logger)
	defer app.Close()

	// 建立 context 用於優雅關閉（第一次 SIGINT/SIGTERM 會觸發取消）
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 執行應用程式（非阻塞），以便在取消後加上 timeout/第二次中斷處理
	runResult := make(chan error, 1)
	go func() {
		runResult <- app.Run(ctx)
	}()

	select {
	case err := <-runResult:
		if err == nil {
			return
		}
		if errors.Is(err, context.Canceled) {
			os.Exit(130)
		}
		log.Fatalf("Failed to execute application: %v", err)
	case <-ctx.Done():
		logger.LogInfo("Shutdown signal received, performing graceful shutdown")

		// 第二次中斷直接強制結束；否則等待 graceful 完成，超時後強退。
		forceSig := make(chan os.Signal, 1)
		signal.Notify(forceSig, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(forceSig)

		select {
		case err := <-runResult:
			if err == nil || errors.Is(err, context.Canceled) {
				os.Exit(130)
			}
			log.Fatalf("Failed to execute application during shutdown: %v", err)
		case <-forceSig:
			logger.LogInfo("Second shutdown signal received, forcing exit")
			os.Exit(1)
		case <-time.After(15 * time.Second):
			logger.LogInfo("Graceful shutdown timeout reached, forcing exit")
			os.Exit(1)
		}
	}
}
