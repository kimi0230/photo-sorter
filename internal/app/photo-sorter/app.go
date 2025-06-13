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
	"photo-sorter/internal/app/photo-sorter/verify"
	"photo-sorter/internal/app/photo-sorter/worker"
	"photo-sorter/internal/pkg/config"
	"photo-sorter/internal/pkg/logger"

	"go.uber.org/zap"
)

// App 主要應用程式結構
type App struct {
	config    *config.Config
	logger    *logger.Logger
	stats     *stats.Stats
	progress  *progress.Progress
	startTime time.Time
}

// NewApp 建立新的應用程式實例
func NewApp(cfg *config.Config, log *logger.Logger) *App {
	return &App{
		config:    cfg,
		logger:    log,
		stats:     stats.NewStats(),
		progress:  progress.NewProgress(),
		startTime: time.Now(),
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
		zap.Bool("是否啟用驗證", a.config.EnableVerify),
		zap.Any("忽略的檔案", a.config.Ignore),
		zap.Any("支援的檔案格式", a.config.Formats),
		zap.String("日期格式", a.config.DateFormat),
		zap.Bool("是否啟用地理位置標籤", a.config.EnableGeoTag),
		zap.String("地理編碼器類型", string(a.config.GeocoderType)),
		zap.String("日誌等級", a.config.LogLevel),
		zap.Bool("是否啟用驗證", a.config.EnableVerify),
		zap.Any("忽略的檔案", a.config.Ignore),
		zap.Any("支援的檔案格式", a.config.Formats),
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
		go func(id int) {
			defer wg.Done()
			worker.Worker(ctx, id, jobs, results, a.config, a.logger, a.progress, a.stats)
		}(i)
	}

	// 發送工作
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
						a.logger.LogInfo("", zap.String("收到取消信號，停止發送工作", ""))
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
	go func() {
		wg.Wait()
		close(results)
	}()

	// 處理結果
	for err := range results {
		if err != nil {
			if err.Error() == "context canceled" {
				a.logger.LogInfo("程式被取消",
					zap.String("status", "canceled"),
				)
				return fmt.Errorf("程式被取消")
			}
			a.logger.LogError("", fmt.Sprintf("處理檔案失敗: %v", err))
		}
	}

	// 輸出統計資訊
	duration := time.Since(startTime)
	stats := a.stats.GetStats()

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

	// 驗證目錄
	matchResult := ""
	if a.config.EnableVerify {
		result, err := verify.CompareDirectories(a.config.SrcDir, a.config.DstDir)
		if err != nil {
			a.logger.LogError("", fmt.Sprintf("驗證目錄失敗: %v", err))
		}
		match := verify.IsMatch(result, a.config.Ignore)
		if match {
			matchResult = "目錄匹配成功"
			a.logger.LogInfo("目錄匹配成功")
		} else {
			matchResult = "目錄不匹配"
			a.logger.LogError("", "目錄不匹配")
		}
	}

	a.logger.LogInfo("處理完成",
		zap.Int("total_files", stats.TotalFiles),
		zap.Int("success_count", stats.SuccessCount),
		zap.Int("failure_count", stats.FailureCount),
		zap.String("result", matchResult),
		zap.Duration("duration", duration),
	)
	fmt.Printf("\n========== 處理完成 ==========\n")
	fmt.Printf("總檔案數: %d\n", stats.TotalFiles)
	fmt.Printf("成功處理: %d\n", stats.SuccessCount)
	fmt.Printf("處理失敗: %d\n", stats.FailureCount)
	fmt.Printf("目錄匹配結果: %s\n", matchResult)
	fmt.Printf("處理時間: %v\n", duration)
	fmt.Printf("========== 處理完成 ==========\n")

	// 檢查是否被取消
	if ctx.Err() != nil {
		a.logger.LogInfo("程式被取消",
			zap.String("status", "canceled"),
			zap.Error(ctx.Err()),
		)
		return fmt.Errorf("程式被取消: %v", ctx.Err())
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
