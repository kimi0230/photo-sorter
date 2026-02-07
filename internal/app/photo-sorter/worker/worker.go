package worker

import (
	"context"
	"fmt"

	"photo-sorter/internal/app/photo-sorter/file"
	"photo-sorter/internal/app/photo-sorter/progress"
	"photo-sorter/internal/app/photo-sorter/stats"
	"photo-sorter/internal/pkg/config"
	"photo-sorter/internal/pkg/logger"
	"photo-sorter/internal/pkg/metadata"
	"photo-sorter/internal/pkg/tagger"

	"go.uber.org/zap"
)

// Worker 處理檔案的工作者
func Worker(ctx context.Context, id int, jobs <-chan string, results chan<- error, cfg *config.Config, logger *logger.Logger, progress *progress.Progress, stats *stats.Stats) {
	var cachedTagger tagger.Tagger
	var cachedTaggerErr error
	exiftoolClient, err := metadata.NewExiftoolClient()
	if err != nil {
		logger.LogError("", fmt.Sprintf("Worker %d failed to start exiftool client: %v", id, err))
	}
	defer func() {
		if exiftoolClient != nil {
			if err := exiftoolClient.Close(); err != nil {
				logger.LogError("", fmt.Sprintf("Worker %d failed to close exiftool client: %v", id, err))
			}
		}
	}()

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
			err := file.ProcessFile(ctx, path, cfg, logger, exiftoolClient, tagProvider)
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
