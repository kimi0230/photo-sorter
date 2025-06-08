package photosorter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"photo-sorter/internal/pkg/config"
	"photo-sorter/internal/pkg/exif"
	"photo-sorter/internal/pkg/logger"

	"go.uber.org/zap"
)

// dirStats 用於統計目錄資訊
type dirStats struct {
	path      string
	fileCount int
	subDirs   map[string]*dirStats
}

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
	targetPath, err := exif.GetTargetPath(path, exifData, a.config)
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
	a.logger.LogInfo("開始處理",
		zap.String("來源資料夾", a.config.SrcDir),
		zap.String("目標資料夾", a.config.DstDir),
	)

	if err := a.printDirectoryStats(a.config.SrcDir); err != nil {
		a.logger.LogError("", fmt.Sprintf("統計資料夾資訊失敗: %v", err))
	}
	// 記錄開始時間
	startTime := time.Now()

	// 建立工作通道
	jobs := make(chan string, 100)
	results := make(chan error, 100)

	// 啟動工作池
	fmt.Printf("啟動工作池，併發數: %d\n", a.config.Workers)
	a.logger.LogInfo("啟動工作池", zap.Int("workers", a.config.Workers))
	var wg sync.WaitGroup
	for i := 0; i < a.config.Workers; i++ {
		wg.Add(1)
		go a.worker(ctx, i, jobs, results, &wg)
	}

	// 使用 goroutine 處理結果
	go func() {
		for err := range results {
			if err != nil {
				a.stats.failureCount++
			} else {
				a.stats.successCount++
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
					return nil
				}

				// 檢查是否為支援的格式
				if a.config.IsSupportedFormat(path) {
					a.stats.totalFiles++
					select {
					case <-ctx.Done():
						return ctx.Err()
					case jobs <- path:
					}
				} else {
					// 處理不支援的檔案
					a.stats.totalFiles++
					a.stats.unsupportedExts[filepath.Ext(path)]++

					// 建立 unknown_format 資料夾
					unknownDir := filepath.Join(a.config.DstDir, "unknown_format")
					if err := os.MkdirAll(unknownDir, 0755); err != nil {
						a.logger.LogError(path, fmt.Sprintf("建立 unknown_format 資料夾失敗: %v", err))
						return err
					}

					// 處理檔案名稱衝突
					baseName := filepath.Base(path)
					targetPath := filepath.Join(unknownDir, baseName)
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
						fmt.Printf("將移動不支援的檔案: %s -> %s\n", path, targetPath)
						a.logger.LogError(path, fmt.Sprintf("將移動不支援的檔案: %s -> %s", path, targetPath))
						return nil
					}

					if err := a.copyFile(path, targetPath); err != nil {
						a.logger.LogError(path, fmt.Sprintf("複製不支援的檔案失敗: %v", err))
						a.stats.failureCount++
					} else {
						a.stats.successCount++
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
	a.logger.LogInfo("處理完成",
		zap.Int("total_files", a.stats.totalFiles),
		zap.Int("success_count", a.stats.successCount),
		zap.Int("failure_count", a.stats.failureCount),
		zap.Duration("duration", duration),
	)
	fmt.Printf("\n處理完成:\n")
	fmt.Printf("總檔案數: %d\n", a.stats.totalFiles)
	fmt.Printf("成功處理: %d\n", a.stats.successCount)
	fmt.Printf("處理失敗: %d\n", a.stats.failureCount)
	fmt.Printf("處理時間: %v\n", duration)

	// 輸出不支援的檔案格式統計
	if len(a.stats.unsupportedExts) > 0 {
		a.logger.LogInfo("不支援的檔案格式統計",
			zap.Any("unsupported_formats", a.stats.unsupportedExts),
		)
	}

	// 統計每個資料夾的檔案數量
	if err := a.printDirectoryStats(a.config.DstDir); err != nil {
		a.logger.LogError("", fmt.Sprintf("統計資料夾資訊失敗: %v", err))
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

func (a *App) calculateTotalFiles(dir *dirStats) int {
	total := dir.fileCount
	for _, subDir := range dir.subDirs {
		total += a.calculateTotalFiles(subDir)
	}
	return total
}

func (a *App) printDirStatsRecursive(dir *dirStats, level int) {
	// 計算總檔案數（包含子目錄）
	totalFiles := a.calculateTotalFiles(dir)

	// 輸出當前目錄資訊
	indent := strings.Repeat("  ", level)
	dirName := filepath.Base(dir.path)
	// if dirName == "." {
	// 	dirName = "sorted_media"
	// }
	a.logger.LogInfo(fmt.Sprintf("%s%s/ (%d 個檔案)", indent, dirName, totalFiles))

	// 遞迴輸出子目錄
	for _, subDir := range dir.subDirs {
		a.printDirStatsRecursive(subDir, level+1)
	}
}

func (a *App) printDirectoryStats(dir string) error {
	root := &dirStats{
		path:    dir,
		subDirs: make(map[string]*dirStats),
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳過根目錄
		if path == dir {
			return nil
		}

		// 取得相對路徑
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		// 分割路徑
		parts := strings.Split(relPath, string(filepath.Separator))
		current := root

		// 建立或更新目錄統計
		for i, part := range parts {
			if i == len(parts)-1 {
				// 這是檔案
				if !info.IsDir() {
					current.fileCount++
				}
			} else {
				// 這是目錄
				if _, exists := current.subDirs[part]; !exists {
					current.subDirs[part] = &dirStats{
						path:    filepath.Join(current.path, part),
						subDirs: make(map[string]*dirStats),
					}
				}
				current = current.subDirs[part]
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	// 輸出統計資訊
	a.logger.LogInfo("資料夾統計資訊")
	a.printDirStatsRecursive(root, 0)

	// 計算實際的總檔案數
	actualTotal := a.calculateTotalFiles(root)
	a.logger.LogInfo(fmt.Sprintf("總計：%d 個檔案", actualTotal))

	return nil
}
