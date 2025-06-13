package worker

import (
	"context"
	"fmt"
	"sync"

	"photo-sorter/internal/app/photo-sorter/file"
	"photo-sorter/internal/app/photo-sorter/progress"
	"photo-sorter/internal/app/photo-sorter/stats"
	"photo-sorter/internal/pkg/config"
	"photo-sorter/internal/pkg/logger"
)

// Worker 處理檔案的工作者
func Worker(ctx context.Context, id int, jobs <-chan string, results chan<- error, wg *sync.WaitGroup, cfg *config.Config, logger *logger.Logger, progress *progress.Progress, stats *stats.Stats) {
	defer wg.Done()
	logger.LogDebug(fmt.Sprintf("Worker %d 已啟動", id))
	for {
		select {
		case <-ctx.Done():
			logger.LogDebug(fmt.Sprintf("Worker %d 已停止", id))
			return
		case path, ok := <-jobs:
			if !ok {
				logger.LogDebug(fmt.Sprintf("Worker %d 已完成所有工作", id))
				return
			}
			logger.LogDebug(fmt.Sprintf("Worker %d 正在處理: %s", id, path))
			progress.Update()
			err := file.ProcessFile(path, cfg, logger)
			if err != nil {
				logger.LogError(path, fmt.Sprintf("Worker %d 處理失敗: %v", id, err))
				// stats.IncrementFailure()
			} else {
				logger.LogDebug(fmt.Sprintf("Worker %d 處理成功: %s", id, path))
				// stats.IncrementSuccess()
			}
			results <- err
		}
	}
}
