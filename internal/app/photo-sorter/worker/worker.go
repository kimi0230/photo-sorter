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

// Result reports a file processing outcome.
type Result struct {
	Path string
	Err  error
}

// Worker 處理檔案的工作者
func Worker(ctx context.Context, id int, jobs <-chan string, results chan<- Result, cfg *config.Config, logger *logger.Logger, progress *progress.Progress, stats *stats.Stats, exifReader metadata.ExifReader, geoResolver metadata.GeoResolver, tagProvider func() (tagger.Tagger, error)) {
	defer func() {
		if exifReader == nil {
			return
		}
		if err := exifReader.Close(); err != nil {
			logger.LogError("", fmt.Sprintf("Worker %d failed to close exif reader: %v", id, err))
		}
	}()

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
			err := file.ProcessFile(ctx, path, cfg, logger, exifReader, geoResolver, tagProvider)
			if err != nil {
				err = fmt.Errorf("worker %d failed for %s: %w", id, path, err)
				stats.IncrementFailure()
			} else {
				logger.LogDebug("Worker succeeded",
					zap.Int("worker_id", id),
					zap.String("path", path),
				)
				stats.IncrementSuccess()
			}
			if err != nil {
				results <- Result{Path: path, Err: err}
			}
		}
	}
}
