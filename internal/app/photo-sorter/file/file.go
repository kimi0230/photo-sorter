package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"photo-sorter/internal/pkg/config"
	"photo-sorter/internal/pkg/logger"
	"photo-sorter/internal/pkg/metadata"
	"photo-sorter/internal/pkg/tagger"

	"go.uber.org/zap"
)

// ProcessFile 處理單個檔案
func ProcessFile(ctx context.Context, path string, cfg *config.Config, logger *logger.Logger, exifReader metadata.ExifReader, geoResolver metadata.GeoResolver, taggerProvider func() (tagger.Tagger, error)) error {
	// 檢查 context 是否已取消
	select {
	case <-ctx.Done():
		return fmt.Errorf("processing canceled: %w", ctx.Err())
	default:
	}

	// 取得 EXIF 資料
	exifData, err := exifReader.GetExifData(path)
	if err != nil {
		logger.LogInfo(path, zap.String("exif_error", "moving file to failed folder"))
		return HandelFailedFolder(path, cfg)
	}

	// 檢查 context 是否已取消
	select {
	case <-ctx.Done():
		return fmt.Errorf("processing canceled: %w", ctx.Err())
	default:
	}

	// 取得目標路徑
	targetPath, location, err := metadata.GetTargetPathWithGeoResolver(path, exifData, cfg, geoResolver)
	if err != nil {
		return fmt.Errorf("failed to get target path: %v", err)
	}

	if cfg.DryRun {
		fmt.Printf("DryRun: move %s -> %s\n", path, targetPath)
		return nil
	}

	// 移動檔案
	if err := MoveFile(path, targetPath); err != nil {
		return fmt.Errorf("failed to move file: %v", err)
	}

	// 檢查 context 是否已取消
	select {
	case <-ctx.Done():
		return fmt.Errorf("processing canceled: %w", ctx.Err())
	default:
	}

	// Add tags when geo tagging is enabled and location is available.
	if cfg.EnableGeoTag && location != nil {
		if taggerProvider == nil {
			return fmt.Errorf("tagger provider is nil")
		}
		if !cfg.DryRun {
			fileTagger, err := taggerProvider()
			if err != nil {
				return fmt.Errorf("failed to create tagger: %v", err)
			}
			tagName := fmt.Sprintf("%s-%s", location.Country, strings.ReplaceAll(location.City, " ", "_"))
			if err := fileTagger.AddTag(targetPath, tagName); err != nil {
				fmt.Printf("Failed to add tag: %v\n", err)
			}
		} else {
			fmt.Printf("DryRun: add tag to %s\n", targetPath)
		}
	}

	return nil
}

// HandleUnsupportedFile 處理不支援的檔案
func HandleUnsupportedFile(path string, cfg *config.Config) error {
	// 建立 unknown_format 資料夾
	unknownDir := filepath.Join(cfg.DstDir, "unknown_format")
	if err := os.MkdirAll(unknownDir, 0755); err != nil {
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

	if cfg.DryRun {
		fmt.Printf("DryRun: move unsupported file %s -> %s\n", path, targetPath)
		return nil
	}

	return MoveFile(path, targetPath)
}

// HandelFailedFolder 將檔案移動到失敗資料夾
func HandelFailedFolder(path string, cfg *config.Config) error {
	// 建立失敗資料夾
	failDir := filepath.Join(cfg.DstDir, "failed_files")

	if err := os.MkdirAll(failDir, 0755); err != nil {
		return err
	}

	// 如果目標檔案已存在，添加時間戳記
	baseName := filepath.Base(path)
	targetPath := filepath.Join(failDir, baseName)
	counter := 1
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)
	for {
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			break
		}
		targetPath = filepath.Join(failDir, fmt.Sprintf("%s_%d%s", nameWithoutExt, counter, ext))
		counter++
	}

	if cfg.DryRun {
		fmt.Printf("DryRun: move failed file %s -> %s\n", path, targetPath)
		return nil
	}
	return MoveFile(path, targetPath)

}

// MoveFile 移動檔案（保留所有時間戳記）
func MoveFile(src, dst string) error {
	// 確保目標目錄存在
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	// 嘗試直接移動（同一個檔案系統內）
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// 如果跨檔案系統移動失敗，則複製後刪除
	// 先複製檔案
	if err := copyFileForMove(src, dst); err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	// 刪除原始檔案
	if err := os.Remove(src); err != nil {
		// 如果刪除失敗，嘗試刪除已複製的檔案
		os.Remove(dst)
		return fmt.Errorf("failed to remove source file: %v", err)
	}

	return nil
}

// copyFileForMove 用於跨檔案系統移動時的複製（保留時間戳記）
func copyFileForMove(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	// 取得原始檔案的資訊
	fileInfo, err := source.Stat()
	if err != nil {
		return err
	}
	modTime := fileInfo.ModTime()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	// 複製內容
	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}

	// 保留原始檔案的修改時間和訪問時間
	if err := os.Chtimes(dst, modTime, modTime); err != nil {
		return err
	}

	return nil
}

// CopyFileWithBuffer 使用 buffer 複製檔案（適合大檔案）
func CopyFileWithBuffer(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	// 取得檔案大小
	fileInfo, err := source.Stat()
	if err != nil {
		return err
	}
	fileSize := fileInfo.Size()

	// 根據檔案大小決定 buffer 大小
	var bufferSize int
	switch {
	case fileSize < 1024*1024: // 小於 1MB
		bufferSize = 32 * 1024 // 32KB
	case fileSize < 10*1024*1024: // 1MB 到 10MB
		bufferSize = 256 * 1024 // 256KB
	case fileSize < 100*1024*1024: // 10MB 到 100MB
		bufferSize = 1024 * 1024 // 1MB
	default: // 大於 100MB
		bufferSize = 4 * 1024 * 1024 // 4MB
	}

	buffer := make([]byte, bufferSize)
	for {
		n, err := source.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		if _, err := destination.Write(buffer[:n]); err != nil {
			return err
		}
	}
	return nil
}

// CopyFileDirect 直接複製檔案（適合小檔案）
func CopyFileDirect(src, dst string) error {
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

// CopyFile 智能選擇的檔案複製函數
func CopyFile(src, dst string) error {
	fileInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// 大檔案使用 Buffer 版本
	if fileInfo.Size() > 100*1024*1024 {
		return CopyFileWithBuffer(src, dst)
	}

	// 小檔案使用直接複製版本
	return CopyFileDirect(src, dst)
}
