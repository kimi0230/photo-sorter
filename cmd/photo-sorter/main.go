package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	photosorter "photo-sorter/internal/app/photo-sorter"
	"photo-sorter/internal/pkg/config"
)

var (
	srcDir     string
	dstDir     string
	workers    int
	dryRun     bool
	configPath string
)

func init() {
	flag.StringVar(&srcDir, "src", ".", "原始照片資料夾")
	flag.StringVar(&dstDir, "dst", ".", "整理後儲存的位置")
	flag.IntVar(&workers, "workers", 4, "最大併發數")
	flag.BoolVar(&dryRun, "dry-run", false, "僅顯示將搬移的路徑，不實際執行")
	flag.StringVar(&configPath, "c", "config/config.yaml", "配置檔案路徑")
}

func main() {
	// 解析命令列參數
	flag.Parse()

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
