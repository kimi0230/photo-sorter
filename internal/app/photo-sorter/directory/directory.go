package directory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"photo-sorter/internal/pkg/logger"
)

// DirStats 用於統計目錄資訊
type DirStats struct {
	path      string
	fileCount int
	subDirs   map[string]*DirStats
}

// NewDirStats 建立新的目錄統計實例
func NewDirStats(path string) *DirStats {
	return &DirStats{
		path:    path,
		subDirs: make(map[string]*DirStats),
	}
}

// CalculateTotalFiles 計算總檔案數
func (d *DirStats) CalculateTotalFiles() int {
	total := d.fileCount
	for _, subDir := range d.subDirs {
		total += subDir.CalculateTotalFiles()
	}
	return total
}

// PrintDirStatsRecursive 遞迴輸出目錄統計
func (d *DirStats) PrintDirStatsRecursive(level int, logger *logger.Logger) {
	// 計算當前目錄的檔案數（不包含子目錄）
	currentDirFiles := d.fileCount

	// 輸出當前目錄資訊
	indent := strings.Repeat("  ", level)
	dirName := filepath.Base(d.path)

	// 如果目錄中有檔案，顯示當前目錄的檔案數
	if currentDirFiles > 0 {
		logger.LogInfo(fmt.Sprintf("%s%s/ (%d 個檔案)", indent, dirName, currentDirFiles))
	} else {
		logger.LogInfo(fmt.Sprintf("%s%s/", indent, dirName))
	}

	// 遞迴輸出子目錄
	for _, subDir := range d.subDirs {
		subDir.PrintDirStatsRecursive(level+1, logger)
	}
}

// PrintDirectoryStats 輸出目錄統計資訊
func PrintDirectoryStats(dir string, logger *logger.Logger) error {
	root := NewDirStats(dir)

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
					current.subDirs[part] = NewDirStats(filepath.Join(current.path, part))
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
	logger.LogInfo("資料夾統計資訊")
	root.PrintDirStatsRecursive(0, logger)

	// 計算實際的總檔案數
	actualTotal := root.CalculateTotalFiles()
	logger.LogInfo(fmt.Sprintf("總計：%d 個檔案", actualTotal))

	return nil
}
