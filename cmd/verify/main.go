package main

import (
	"flag"
	"fmt"
	"os"

	"photo-sorter/internal/app/photo-sorter/verify"
)

func main() {
	// 定義命令列參數
	sourceDir := flag.String("source", "", "來源目錄路徑")
	targetDir := flag.String("target", "", "目標目錄路徑")
	flag.Parse()

	// 檢查必要參數
	if *sourceDir == "" || *targetDir == "" {
		fmt.Println("錯誤：需要提供來源和目標目錄路徑")
		fmt.Println("使用方式：verify -source <來源目錄> -target <目標目錄>")
		os.Exit(1)
	}

	// 比對目錄
	result, err := verify.CompareDirectories(*sourceDir, *targetDir)
	if err != nil {
		fmt.Printf("錯誤：%v\n", err)
		os.Exit(1)
	}

	// 印出結果
	fmt.Printf("比對目錄：\n")
	fmt.Printf("來源目錄：%s\n", *sourceDir)
	fmt.Printf("目標目錄：%s\n", *targetDir)
	verify.PrintResult(result)
}
