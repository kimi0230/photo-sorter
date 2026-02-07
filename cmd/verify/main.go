package main

import (
	"flag"
	"os"

	"photo-sorter/internal/app/photosorter/verify"
	"photo-sorter/internal/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	// 定義命令列參數
	sourceDir := flag.String("source", "", "Source directory path")
	targetDir := flag.String("target", "", "Target directory path")
	flag.Parse()

	log, err := logger.NewLogger("info")
	if err != nil {
		panic(err)
	}
	defer log.Close()

	// 檢查必要參數
	if *sourceDir == "" || *targetDir == "" {
		log.LogWarn("Missing required parameters",
			zap.String("source", *sourceDir),
			zap.String("target", *targetDir),
		)
		log.LogInfo("Usage",
			zap.String("example", "verify -source <Source directory> -target <Target directory>"),
		)
		os.Exit(1)
	}

	// 比對目錄
	result, err := verify.CompareDirectories(*sourceDir, *targetDir)
	if err != nil {
		log.LogError("", err.Error())
		os.Exit(1)
	}

	// 印出結果
	log.LogInfo("Compare directories",
		zap.String("source", *sourceDir),
		zap.String("target", *targetDir),
	)
	verify.PrintResult(result, log)
}
