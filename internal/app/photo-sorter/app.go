package photosorter

import (
	"context"
	"errors"
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
	a.logger.LogInfo("Start processing",
		zap.String("source_dir", a.config.SrcDir),
		zap.String("dest_dir", a.config.DstDir),
		zap.Bool("enable_verify", a.config.EnableVerify),
		zap.Any("ignored_files", a.config.Ignore),
		zap.Any("supported_formats", a.config.Formats),
		zap.String("date_format", a.config.DateFormat),
		zap.Bool("enable_geo_tag", a.config.EnableGeoTag),
		zap.String("geocoder_type", string(a.config.GeocoderType)),
		zap.String("log_level", a.config.LogLevel),
	)

	if err := directory.PrintDirectoryStats(a.config.SrcDir, a.logger); err != nil {
		a.logger.LogError("", fmt.Sprintf("Failed to collect directory stats: %v", err))
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

	jobList, totalFiles, ignoredFiles, err := a.scanJobs()
	if err != nil {
		return fmt.Errorf("failed to count total files: %v", err)
	}

	// 設定總檔案數
	a.progress.SetTotal(totalFiles)
	a.stats.SetTotalFiles(totalFiles)

	// 啟動工作池
	fmt.Printf("Workers: %d, total files: %d, ignored files: %d\n", a.config.Workers, totalFiles, ignoredFiles)
	a.logger.LogInfo("Start Workers",
		zap.Int("workers", a.config.Workers),
		zap.Int("total_files", totalFiles),
	)

	wg := a.startWorkers(ctx, jobs, results)

	// Dispatch jobs based on the scanned list.
	go func() {
		defer close(jobs)
		for _, path := range jobList {
			select {
			case <-ctx.Done():
				a.logger.LogInfo("Cancel signal received, stop dispatching jobs")
				return
			case jobs <- path:
			}
		}
	}()

	// 等待所有工作完成
	go func() {
		wg.Wait()
		close(results)
	}()

	if err := a.collectResults(ctx, results); err != nil {
		return err
	}

	// 輸出統計資訊
	duration := time.Since(startTime)
	stats := a.stats.GetStats()

	// 輸出不支援的檔案格式統計
	if len(stats.UnsupportedExts) > 0 {
		a.logger.LogInfo("Unsupported format stats",
			zap.Any("unsupported_formats", stats.UnsupportedExts),
		)
	}

	// 輸出被忽略的檔案格式統計
	if len(stats.IgnoredExts) > 0 {
		a.logger.LogInfo("Ignored format stats",
			zap.Any("ignored_formats", stats.IgnoredExts),
		)
	}

	// 統計每個資料夾的檔案數量
	if err := directory.PrintDirectoryStats(a.config.DstDir, a.logger); err != nil {
		a.logger.LogError("", fmt.Sprintf("Failed to collect directory stats: %v", err))
	}

	// 驗證目錄（改用移動後，只需檢查檔案數量）
	matchResult := ""
	if a.config.EnableVerify {
		// 計算來源目錄剩餘檔案數（應該只剩下失敗的檔案）
		sourceCount := 0
		err := filepath.Walk(a.config.SrcDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			// 排除目標目錄
			if strings.HasPrefix(path, a.config.DstDir) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if !info.IsDir() {
				// 排除忽略的檔案
				if !a.config.ShouldIgnore(path) {
					sourceCount++
				}
			}
			return nil
		})
		if err != nil {
			a.logger.LogError("", fmt.Sprintf("Failed to count source files: %v", err))
		}

		// 計算目標目錄檔案數（包含成功處理的檔案和失敗的檔案）
		targetCount := 0
		err = filepath.Walk(a.config.DstDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() {
				targetCount++
			}
			return nil
		})
		if err != nil {
			a.logger.LogError("", fmt.Sprintf("Failed to count target files: %v", err))
		}

		// 驗證邏輯：
		// 1. 目標目錄檔案數應該等於成功處理數 + 失敗處理數（失敗的檔案會移到 failed_files）
		// 2. 來源目錄應該只剩下未處理的檔案（如果有）
		expectedTargetCount := stats.SuccessCount + stats.FailureCount

		// 允許一些誤差（因為可能有其他檔案，如系統檔案等）
		// 主要檢查目標目錄的檔案數是否合理
		if targetCount >= expectedTargetCount && sourceCount <= stats.TotalFiles-expectedTargetCount {
			matchResult = "directory match succeeded"
			a.logger.LogInfo("Directory match succeeded",
				zap.Int("source_remaining_files", sourceCount),
				zap.Int("target_files", targetCount),
				zap.Int("success_count", stats.SuccessCount),
				zap.Int("failure_count", stats.FailureCount),
				zap.Int("expected_target_files", expectedTargetCount),
			)
		} else {
			matchResult = "directory match failed"
			a.logger.LogInfo("Directory match failed",
				zap.Int("source_remaining_files", sourceCount),
				zap.Int("target_files", targetCount),
				zap.Int("success_count", stats.SuccessCount),
				zap.Int("failure_count", stats.FailureCount),
				zap.Int("expected_target_files", expectedTargetCount),
			)
		}
	}

	a.logger.LogInfo("Processing completed",
		zap.Int("total_files", stats.TotalFiles),
		zap.Int("success_count", stats.SuccessCount),
		zap.Int("failure_count", stats.FailureCount),
		zap.String("result", matchResult),
		zap.Duration("duration", duration),
	)
	fmt.Printf("\n========== Processing Completed ==========\n")
	fmt.Printf("Total files: %d\n", stats.TotalFiles)
	fmt.Printf("Succeeded: %d\n", stats.SuccessCount)
	fmt.Printf("Failed: %d\n", stats.FailureCount)
	fmt.Printf("Directory match: %s\n", matchResult)
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("========== Processing Completed ==========\n")

	// 檢查是否被取消
	if ctx.Err() != nil {
		a.logger.LogInfo("Process canceled",
			zap.String("status", "canceled"),
			zap.Error(ctx.Err()),
		)
		return fmt.Errorf("process canceled: %v", ctx.Err())
	}

	return nil
}

func (a *App) scanJobs() ([]string, int, int, error) {
	totalFiles, ignoredFiles := 0, 0
	jobList := make([]string, 0, 1024)
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
				a.logger.LogInfo(path, zap.String("ignored_file_ext", filepath.Ext(path)))
				a.stats.IncrementIgnoredExt(filepath.Ext(path))
				ignoredFiles++
				return nil
			}
			totalFiles++

			// 檢查是否為支援的格式
			if a.config.IsSupportedFormat(path) {
				jobList = append(jobList, path)
			} else {
				// 處理不支援的檔案
				a.stats.IncrementUnsupportedExt(filepath.Ext(path))
				if err := file.HandleUnsupportedFile(path, a.config, a.logger); err != nil {
					a.logger.LogError(path, fmt.Sprintf("Failed to handle unsupported file: %v", err))
					a.stats.IncrementFailure()
				} else {
					a.logger.LogDebug(path, zap.String("unsupported_file_handled", filepath.Ext(path)))
					a.stats.IncrementSuccess()
				}
			}
		}
		return nil
	})
	return jobList, totalFiles, ignoredFiles, err
}

func (a *App) startWorkers(ctx context.Context, jobs <-chan string, results chan<- error) *sync.WaitGroup {
	var wg sync.WaitGroup
	for i := 0; i < a.config.Workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			worker.Worker(ctx, id, jobs, results, a.config, a.logger, a.progress, a.stats)
		}(i)
	}
	return &wg
}

func (a *App) collectResults(ctx context.Context, results <-chan error) error {
	for err := range results {
		if err != nil {
			if errors.Is(err, context.Canceled) {
				a.logger.LogInfo("Process canceled",
					zap.String("status", "canceled"),
				)
				return fmt.Errorf("process canceled")
			}
			a.logger.LogError("", fmt.Sprintf("Failed to process file: %v", err))
		}
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
				fmt.Printf("\rProgress: %.1f%% (%d/%d)\n",
					percentage, processed, total)
			}
		}
	}
}
