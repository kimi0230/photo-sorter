package worker

import (
	"context"
	"fmt"

	"photo-sorter/internal/app/photo-sorter/file"
	"photo-sorter/internal/app/photo-sorter/progress"
	"photo-sorter/internal/app/photo-sorter/stats"
	"photo-sorter/internal/pkg/config"
	"photo-sorter/internal/pkg/logger"
	"photo-sorter/internal/pkg/tagger"

	"go.uber.org/zap"
)

// Worker 處理檔案的工作者
func Worker(ctx context.Context, id int, jobs <-chan string, results chan<- error, cfg *config.Config, logger *logger.Logger, progress *progress.Progress, stats *stats.Stats) {
	var cachedTagger tagger.Tagger
	var cachedTaggerErr error
	tagProvider := func() (tagger.Tagger, error) {
		if cachedTagger != nil || cachedTaggerErr != nil {
			return cachedTagger, cachedTaggerErr
		}
		cachedTagger, cachedTaggerErr = tagger.NewTagger()
		return cachedTagger, cachedTaggerErr
	}

	for path := range jobs {
		select {
		case <-ctx.Done():
			logger.LogDebug("Worker received cancel signal",
				zap.Int("worker_id", id),
				zap.String("status", "stopped"),
			)
			return
		default:
			logger.LogDebug("Worker processing file",
				zap.Int("worker_id", id),
				zap.String("path", path),
			)
			progress.Update()
			err := file.ProcessFile(ctx, path, cfg, logger, tagProvider)
			if err != nil {
				logger.LogError(path, fmt.Sprintf("Worker %d failed: %v", id, err))
				stats.IncrementFailure()
			} else {
				logger.LogDebug("Worker succeeded",
					zap.Int("worker_id", id),
					zap.String("path", path),
				)
				stats.IncrementSuccess()
			}
			results <- err
		}
	}
}
