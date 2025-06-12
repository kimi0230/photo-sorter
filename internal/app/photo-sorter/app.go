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

// Progress 用於追蹤處理進度
type Progress struct {
	mu             sync.RWMutex
	processedFiles int
	totalFiles     int
}

func (p *Progress) Update() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.processedFiles++
}

func (p *Progress) SetTotal(total int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.totalFiles = total
}

func (p *Progress) GetStatus() (processed, total int) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.processedFiles, p.totalFiles
}

// Stats 用於追蹤處理統計
type Stats struct {
	totalFiles      int
	successCount    int
	failureCount    int
	unsupportedExts map[string]int
}

type App struct {
	config   *config.Config
	logger   *logger.Logger
	stats    Stats
	progress *Progress
	mu       sync.Mutex
}

func NewApp(cfg *config.Config) (*App, error) {
	log, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}

	return &App{
		config: cfg,
		logger: log,
		stats: Stats{
			unsupportedExts: make(map[string]int),
		},
		progress: &Progress{},
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
	a.setTotalFiles(totalFiles)

	// 啟動工作池
	fmt.Printf("Workers 數量: %d，需處理總檔案數: %d，忽略的檔案: %d\n", a.config.Workers, totalFiles, ignoredFiles)
	a.logger.LogInfo("Start Workers",
		zap.Int("workers", a.config.Workers),
		zap.Int("total_files", totalFiles),
	)
	var wg sync.WaitGroup
	for i := 0; i < a.config.Workers; i++ {
		wg.Add(1)
		go a.worker(ctx, i, jobs, results, &wg)
	}

	// 使用 goroutine 處理結果
	go func() {
		for err := range results {
			if err != nil {
				a.incrementFailure()
			} else {
				a.incrementSuccess()
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
					select {
					case <-ctx.Done():
						return ctx.Err()
					case jobs <- path:
					}
				} else {
					// 處理不支援的檔案
					a.incrementUnsupportedExt(filepath.Ext(path))

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
						a.incrementFailure()
					} else {
						a.incrementSuccess()
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
	stats := a.getStats()
	a.logger.LogInfo("處理完成",
		zap.Int("total_files", stats.totalFiles),
		zap.Int("success_count", stats.successCount),
		zap.Int("failure_count", stats.failureCount),
		zap.Int("ignored_files", ignoredFiles),
		zap.Duration("duration", duration),
	)
	fmt.Printf("\n處理完成:\n")
	fmt.Printf("總檔案數: %d\n", stats.totalFiles)
	fmt.Printf("成功處理: %d\n", stats.successCount)
	fmt.Printf("處理失敗: %d\n", stats.failureCount)
	fmt.Printf("忽略的檔案: %d\n", ignoredFiles)
	fmt.Printf("處理時間: %v\n", duration)

	// 輸出不支援的檔案格式統計
	if len(stats.unsupportedExts) > 0 {
		a.logger.LogInfo("不支援的檔案格式統計",
			zap.Any("unsupported_formats", stats.unsupportedExts),
		)
	}

	// 統計每個資料夾的檔案數量
	if err := a.printDirectoryStats(a.config.DstDir); err != nil {
		a.logger.LogError("", fmt.Sprintf("統計資料夾資訊失敗: %v", err))
	}

	return nil
}

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
			a.progress.Update()
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
	// 計算當前目錄的檔案數（不包含子目錄）
	currentDirFiles := dir.fileCount

	// 輸出當前目錄資訊
	indent := strings.Repeat("  ", level)
	dirName := filepath.Base(dir.path)

	// 如果目錄中有檔案，顯示當前目錄的檔案數
	if currentDirFiles > 0 {
		a.logger.LogInfo(fmt.Sprintf("%s%s/ (%d 個檔案)", indent, dirName, currentDirFiles))
	} else {
		a.logger.LogInfo(fmt.Sprintf("%s%s/", indent, dirName))
	}

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

// 統計相關的方法
func (a *App) incrementSuccess() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.stats.successCount++
}

func (a *App) incrementFailure() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.stats.failureCount++
}

func (a *App) incrementUnsupportedExt(ext string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.stats.unsupportedExts[ext]++
}

func (a *App) setTotalFiles(total int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.stats.totalFiles = total
}

func (a *App) getStats() Stats {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.stats
}
