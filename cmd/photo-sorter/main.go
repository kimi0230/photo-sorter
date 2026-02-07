package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"

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

	// 建立 context 用於優雅關閉
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 處理信號
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.LogInfo("Shutdown signal received, performing graceful shutdown")
		cancel()
	}()

	// 執行應用程式
	if err := app.Run(ctx); err != nil {
		log.Fatalf("Failed to execute application: %v", err)
	}
}
