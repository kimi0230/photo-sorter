package photosorter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"photo-sorter/internal/app/photo-sorter/directory"
	"photo-sorter/internal/app/photo-sorter/file"
	"photo-sorter/internal/app/photo-sorter/progress"
	"photo-sorter/internal/app/photo-sorter/stats"
	"photo-sorter/internal/app/photo-sorter/worker"
	"photo-sorter/internal/pkg/config"
	"photo-sorter/internal/pkg/logger"

	"go.uber.org/zap"
)

// App 主要應用程式結構
type App struct {
	config   *config.Config
	logger   *logger.Logger
	stats    *stats.Stats
	progress *progress.Progress
}

// NewApp 建立新的應用程式實例
func NewApp(cfg *config.Config, log *logger.Logger) *App {
	return &App{
		config:   cfg,
		logger:   log,
		stats:    stats.NewStats(),
		progress: progress.NewProgress(),
	}
}

// Close 關閉應用程式
func (a *App) Close() error {
	return a.logger.Close()
}

// Run 執行應用程式
func (a *App) Run(ctx context.Context) error {
	a.logger.LogInfo("開始處理",
		zap.String("來源資料夾", a.config.SrcDir),
		zap.String("目標資料夾", a.config.DstDir),
	)

	if err := directory.PrintDirectoryStats(a.config.SrcDir, a.logger); err != nil {
		a.logger.LogError("", fmt.Sprintf("統計資料夾資訊失敗: %v", err))
	}

	// 啟動進度監控
	progressCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go a.monitorProgress(progressCtx)

	// 記錄開始時間
	startTime := time.Now()

	// 建立工作通道
	jobs := make(chan string, 100)
	results := make(chan error, 100)

	// 先計算總檔案數
	totalFiles, ignoredFiles := 0, 0
	err := filepath.Walk(a.config.SrcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 檢查是否為目標目錄或其子目錄
		if strings.HasPrefix(path, a.config.DstDir) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 只計算來源目錄中的檔案
		if !info.IsDir() {
			// 檢查是否要忽略此檔案
			if a.config.ShouldIgnore(path) {
				a.logger.LogInfo(path, zap.String("忽略的檔案", filepath.Ext(path)))
				ignoredFiles++
				return nil
			}
			totalFiles++
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("計算總檔案數失敗: %v", err)
	}

	// 設定總檔案數
	a.progress.SetTotal(totalFiles)
	a.stats.SetTotalFiles(totalFiles)

	// 啟動工作池
	fmt.Printf("Workers 數量: %d，需處理總檔案數: %d，忽略的檔案數: %d\n", a.config.Workers, totalFiles, ignoredFiles)
	a.logger.LogInfo("Start Workers",
		zap.Int("workers", a.config.Workers),
		zap.Int("total_files", totalFiles),
	)

	var wg sync.WaitGroup
	for i := 0; i < a.config.Workers; i++ {
		wg.Add(1)
		go worker.Worker(ctx, i, jobs, results, &wg, a.config, a.logger, a.progress, a.stats)
	}

	// 使用 goroutine 處理結果
	go func() {
		for err := range results {
			if err != nil {
				a.logger.LogError(err.Error(), fmt.Sprintf("處理失敗 %s", err.Error()))
			} else {
				a.logger.LogDebug("處理成功")
			}
		}
	}()

	// 使用 goroutine 掃描檔案並發送工作
	go func() {
		defer close(jobs)
		err := filepath.Walk(a.config.SrcDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// 檢查是否為目標目錄或其子目錄
			if strings.HasPrefix(path, a.config.DstDir) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if !info.IsDir() {
				// 檢查是否要忽略此檔案
				if a.config.ShouldIgnore(path) {
					a.stats.IncrementIgnoredExt(filepath.Ext(path))
					return nil
				}

				// 檢查是否為支援的格式
				if a.config.IsSupportedFormat(path) {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case jobs <- path:
					}
				} else {
					// 處理不支援的檔案
					a.stats.IncrementUnsupportedExt(filepath.Ext(path))
					if err := file.HandleUnsupportedFile(path, a.config, a.logger); err != nil {
						a.logger.LogError(path, fmt.Sprintf("處理不支援的檔案失敗: %v", err))
						a.stats.IncrementFailure()
					} else {
						a.logger.LogDebug(path, zap.String("處理不支援的檔案成功", filepath.Ext(path)))
						a.stats.IncrementSuccess()
					}
				}
			}
			return nil
		})
		if err != nil {
			fmt.Printf("掃描檔案時發生錯誤: %v\n", err)
		}
	}()

	// 等待所有工作完成
	wg.Wait()
	close(results)

	// 計算處理時間
	duration := time.Since(startTime)

	// 輸出統計資訊
	stats := a.stats.GetStats()
	a.logger.LogInfo("處理完成",
		zap.Int("total_files", stats.TotalFiles),
		zap.Int("success_count", stats.SuccessCount),
		zap.Int("failure_count", stats.FailureCount),
		zap.Duration("duration", duration),
	)
	fmt.Printf("\n========== 處理完成 ==========\n")
	fmt.Printf("總檔案數: %d\n", stats.TotalFiles)
	fmt.Printf("成功處理: %d\n", stats.SuccessCount)
	fmt.Printf("處理失敗: %d\n", stats.FailureCount)
	fmt.Printf("處理時間: %v\n", duration)
	fmt.Printf("\n========== 處理完成 ==========\n")

	// 輸出不支援的檔案格式統計
	if len(stats.UnsupportedExts) > 0 {
		a.logger.LogInfo("不支援的檔案格式統計",
			zap.Any("unsupported_formats", stats.UnsupportedExts),
		)
	}

	// 輸出被忽略的檔案格式統計
	if len(stats.IgnoredExts) > 0 {
		a.logger.LogInfo("被忽略的檔案格式統計",
			zap.Any("ignored_formats", stats.IgnoredExts),
		)
	}

	// 統計每個資料夾的檔案數量
	if err := directory.PrintDirectoryStats(a.config.DstDir, a.logger); err != nil {
		a.logger.LogError("", fmt.Sprintf("統計資料夾資訊失敗: %v", err))
	}

	return nil
}

// monitorProgress 監控處理進度
func (a *App) monitorProgress(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			processed, total := a.progress.GetStatus()
			if total > 0 {
				percentage := float64(processed) / float64(total) * 100
				fmt.Printf("\r進度: %.1f%% (%d/%d)\n",
					percentage, processed, total)
			}
		}
	}
}
