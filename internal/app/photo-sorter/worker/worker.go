package worker

import (
	"context"
	"fmt"

	"photo-sorter/internal/app/photo-sorter/file"
	"photo-sorter/internal/app/photo-sorter/progress"
	"photo-sorter/internal/app/photo-sorter/stats"
	"photo-sorter/internal/pkg/config"
	"photo-sorter/internal/pkg/logger"

	"go.uber.org/zap"
)

// Worker 處理檔案的工作者
func Worker(ctx context.Context, id int, jobs <-chan string, results chan<- error, cfg *config.Config, logger *logger.Logger, progress *progress.Progress, stats *stats.Stats) {
	for path := range jobs {
		select {
		case <-ctx.Done():
			logger.LogDebug("Worker 收到取消信號",
				zap.Int("worker_id", id),
				zap.String("status", "stopped"),
			)
			return
		default:
			logger.LogDebug("Worker 正在處理檔案",
				zap.Int("worker_id", id),
				zap.String("path", path),
			)
			progress.Update()
			err := file.ProcessFile(ctx, path, cfg, logger)
			if err != nil {
				logger.LogError(path, fmt.Sprintf("Worker %d 處理失敗: %v", id, err))
				stats.IncrementFailure()
			} else {
				logger.LogDebug("Worker 處理成功",
					zap.Int("worker_id", id),
					zap.String("path", path),
				)
				stats.IncrementSuccess()
			}
			results <- err
		}
	}
}
