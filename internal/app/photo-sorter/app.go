package photosorter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"photo-sorter/internal/pkg/config"
	"photo-sorter/internal/pkg/exif"
	"photo-sorter/pkg/logger"
)

type App struct {
	config *config.Config
	logger *logger.Logger
	stats  struct {
		totalFiles      int
		successCount    int
		failureCount    int
		unsupportedExts map[string]int // 記錄不支援的副檔名及其數量
	}
}

func NewApp(cfg *config.Config) (*App, error) {
	log, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}

	return &App{
		config: cfg,
		logger: log,
		stats: struct {
			totalFiles      int
			successCount    int
			failureCount    int
			unsupportedExts map[string]int
		}{
			unsupportedExts: make(map[string]int),
		},
	}, nil
}

func (a *App) Close() error {
	return a.logger.Close()
}

func (a *App) ProcessFile(path string) error {
	// 取得檔案資訊
	exifData, err := exif.GetExifData(path)
	if err != nil {
		a.logger.LogError(path, fmt.Sprintf("取得檔案資訊失敗: %v", err))
		return err
	}

	// 取得目標路徑
	targetPath, err := exif.GetTargetPath(path, exifData, a.config.DstDir)
	if err != nil {
		a.logger.LogError(path, fmt.Sprintf("取得目標路徑失敗: %v", err))
		return err
	}

	if a.config.DryRun {
		fmt.Printf("將移動: %s -> %s\n", path, targetPath)
		return nil
	}

	// 複製檔案
	if err := a.copyFile(path, targetPath); err != nil {
		a.logger.LogError(path, fmt.Sprintf("複製檔案失敗: %v", err))
		return err
	}

	return nil
}

func (a *App) copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	err = os.WriteFile(dst, input, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) Run(ctx context.Context) error {
	// 先收集所有需要處理的檔案
	var files []string
	var unsupportedFiles []string
	err := filepath.Walk(a.config.SrcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := filepath.Ext(path)
			switch ext {
			case ".jpg", ".jpeg", ".heic", ".png", ".mp4", ".mov":
				files = append(files, path)
				a.stats.totalFiles++
			default:
				unsupportedFiles = append(unsupportedFiles, path)
				a.stats.unsupportedExts[ext]++
				a.stats.totalFiles++
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("掃描檔案時發生錯誤: %v", err)
	}

	fmt.Printf("找到 %d 個檔案需要處理\n", len(files))
	if len(unsupportedFiles) > 0 {
		fmt.Printf("其中 %d 個檔案為不支援的格式\n", len(unsupportedFiles))
	}

	// 建立工作通道
	jobs := make(chan string, len(files))
	results := make(chan error, len(files))

	// 啟動工作池
	fmt.Printf("啟動工作池，併發數: %d\n", a.config.Workers)
	var wg sync.WaitGroup
	for i := 0; i < a.config.Workers; i++ {
		wg.Add(1)
		go a.worker(ctx, i, jobs, results, &wg)
	}

	// 發送工作
	go func() {
		defer close(jobs)
		// 先處理支援的檔案
		for _, file := range files {
			select {
			case <-ctx.Done():
				return
			case jobs <- file:
			}
		}
	}()

	// 處理結果
	go func() {
		for err := range results {
			if err != nil {
				a.stats.failureCount++
			} else {
				a.stats.successCount++
			}
		}
	}()

	// 等待所有工作完成
	wg.Wait()
	close(results)

	// 處理不支援的檔案
	if len(unsupportedFiles) > 0 {
		fmt.Printf("\n開始處理不支援的檔案格式...\n")
		unknownDir := filepath.Join(a.config.DstDir, "unknown_format")
		if err := os.MkdirAll(unknownDir, 0755); err != nil {
			return fmt.Errorf("建立 unknown_format 資料夾失敗: %v", err)
		}

		for _, file := range unsupportedFiles {
			baseName := filepath.Base(file)
			targetPath := filepath.Join(unknownDir, baseName)

			// 處理檔案名稱衝突
			counter := 1
			ext := filepath.Ext(baseName)
			nameWithoutExt := strings.TrimSuffix(baseName, ext)
			for {
				if _, err := os.Stat(targetPath); os.IsNotExist(err) {
					break
				}
				targetPath = filepath.Join(unknownDir, fmt.Sprintf("%s_%d%s", nameWithoutExt, counter, ext))
				counter++
			}

			if a.config.DryRun {
				fmt.Printf("將移動不支援的檔案: %s -> %s\n", file, targetPath)
				continue
			}

			if err := a.copyFile(file, targetPath); err != nil {
				a.logger.LogError(file, fmt.Sprintf("複製不支援的檔案失敗: %v", err))
				a.stats.failureCount++
			} else {
				a.stats.successCount++
			}
		}
	}

	// 輸出統計資訊
	fmt.Printf("\n處理完成:\n")
	fmt.Printf("總檔案數: %d\n", a.stats.totalFiles)
	fmt.Printf("成功處理: %d\n", a.stats.successCount)
	fmt.Printf("處理失敗: %d\n", a.stats.failureCount)

	// 輸出不支援的檔案格式統計
	if len(a.stats.unsupportedExts) > 0 {
		fmt.Printf("\n不支援的檔案格式:\n")
		for ext, count := range a.stats.unsupportedExts {
			fmt.Printf("%s: %d 個檔案\n", ext, count)
		}
	}

	return nil
}

func (a *App) worker(ctx context.Context, id int, jobs <-chan string, results chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case path, ok := <-jobs:
			if !ok {
				return
			}
			err := a.ProcessFile(path)
			results <- err
		}
	}
}
