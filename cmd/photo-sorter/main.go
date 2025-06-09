package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	photosorter "photo-sorter/internal/app/photo-sorter"
	"photo-sorter/internal/pkg/config"
	"photo-sorter/internal/pkg/version"
)

var (
	srcDir     string
	dstDir     string
	workers    int
	dryRun     bool
	configPath string
	showVer    bool
	cpuProfile string // CPU profile 檔案路徑
	memProfile string // 記憶體 profile 檔案路徑
)

func init() {
	flag.StringVar(&srcDir, "src", ".", "原始照片資料夾")
	flag.StringVar(&dstDir, "dst", ".", "整理後儲存的位置")
	flag.IntVar(&workers, "workers", 4, "最大併發數")
	flag.BoolVar(&dryRun, "dry-run", false, "僅顯示將搬移的路徑，不實際執行")
	flag.StringVar(&configPath, "c", "config.yaml", "配置檔案路徑")
	flag.BoolVar(&showVer, "version", false, "顯示版本資訊")
	flag.StringVar(&cpuProfile, "cpuprofile", "", "CPU profile 檔案路徑")
	flag.StringVar(&memProfile, "memprofile", "", "記憶體 profile 檔案路徑")
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
			log.Fatalf("建立 CPU profile 失敗: %v", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatalf("啟動 CPU profile 失敗: %v", err)
		}
		defer pprof.StopCPUProfile()
	}

	// 啟動記憶體 profiling
	if memProfile != "" {
		f, err := os.Create(memProfile)
		if err != nil {
			log.Fatalf("建立記憶體 profile 失敗: %v", err)
		}
		defer f.Close()
		defer pprof.WriteHeapProfile(f)
	}

	// 載入配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("載入設定檔失敗: %v", err)
	}

	// 套用命令列參數
	cfg.ApplyFlags(srcDir, dstDir, workers, dryRun)

	// 建立應用程式
	app, err := photosorter.NewApp(cfg)
	if err != nil {
		log.Fatalf("建立應用程式失敗: %v", err)
	}
	defer app.Close()

	// 建立 context 用於優雅關閉
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 處理信號
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n收到關閉信號，正在優雅關閉...")
		cancel()
	}()

	// 執行應用程式
	if err := app.Run(ctx); err != nil {
		log.Fatalf("執行應用程式失敗: %v", err)
	}
}
