package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"photo-sorter/internal/pkg/config"
	"photo-sorter/internal/pkg/exif"
	"photo-sorter/internal/pkg/logger"
)

// ProcessFile 處理單個檔案
func ProcessFile(path string, cfg *config.Config, logger *logger.Logger) error {
	// 取得檔案資訊
	exifData, err := exif.GetExifData(path)
	if err != nil {
		logger.LogError(path, fmt.Sprintf("取得檔案資訊失敗: %v", err))
		return err
	}

	// 取得目標路徑
	targetPath, err := exif.GetTargetPath(path, exifData, cfg)
	if err != nil {
		logger.LogError(path, fmt.Sprintf("取得目標路徑失敗: %v", err))
		return err
	}

	if cfg.DryRun {
		fmt.Printf("將移動: %s -> %s\n", path, targetPath)
		return nil
	}

	// 複製檔案
	if err := CopyFile(path, targetPath); err != nil {
		logger.LogError(path, fmt.Sprintf("複製檔案失敗: %v", err))
		return err
	}

	return nil
}

// CopyFile 複製檔案
func CopyFile(src, dst string) error {
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

// HandleUnsupportedFile 處理不支援的檔案
func HandleUnsupportedFile(path string, cfg *config.Config, logger *logger.Logger) error {
	// 建立 unknown_format 資料夾
	unknownDir := filepath.Join(cfg.DstDir, "unknown_format")
	if err := os.MkdirAll(unknownDir, 0755); err != nil {
		logger.LogError(path, fmt.Sprintf("建立 unknown_format 資料夾失敗: %v", err))
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
		fmt.Printf("將移動不支援的檔案: %s -> %s\n", path, targetPath)
		logger.LogError(path, fmt.Sprintf("將移動不支援的檔案: %s -> %s", path, targetPath))
		return nil
	}

	return CopyFile(path, targetPath)
}
