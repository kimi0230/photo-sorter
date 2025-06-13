package verify

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// CompareResult 儲存比對結果
type CompareResult struct {
	OnlyInSource []string
	OnlyInTarget []string
}

// CompareDirectories 比對兩個目錄中的檔案
func CompareDirectories(sourceDir, targetDir string) (*CompareResult, error) {
	// 檢查目錄是否存在
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("來源目錄 '%s' 不存在", sourceDir)
	}
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("目標目錄 '%s' 不存在", targetDir)
	}

	// 取得兩個目錄的檔案列表
	sourceFiles, err := getFileList(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("讀取來源目錄失敗: %v", err)
	}

	targetFiles, err := getFileList(targetDir)
	if err != nil {
		return nil, fmt.Errorf("讀取目標目錄失敗: %v", err)
	}

	// 比對檔案
	result := &CompareResult{
		OnlyInSource: make([]string, 0),
		OnlyInTarget: make([]string, 0),
	}

	// 建立檔案名稱到完整路徑的映射
	sourceMap := make(map[string]string)
	for _, file := range sourceFiles {
		sourceMap[filepath.Base(file)] = file
	}

	targetMap := make(map[string]string)
	for _, file := range targetFiles {
		targetMap[filepath.Base(file)] = file
	}

	// 找出只在來源目錄存在的檔案
	for name := range sourceMap {
		if _, exists := targetMap[name]; !exists {
			result.OnlyInSource = append(result.OnlyInSource, name)
		}
	}

	// 找出只在目標目錄存在的檔案
	for name := range targetMap {
		if _, exists := sourceMap[name]; !exists {
			result.OnlyInTarget = append(result.OnlyInTarget, name)
		}
	}

	// 排序結果
	sort.Strings(result.OnlyInSource)
	sort.Strings(result.OnlyInTarget)

	return result, nil
}

// IsMatch 檢查兩個目錄是否匹配，可以指定要忽略的檔案
func IsMatch(result *CompareResult, ignorePatterns []string) bool {
	// 如果沒有差異，直接返回 true
	if len(result.OnlyInSource) == 0 && len(result.OnlyInTarget) == 0 {
		return true
	}

	// 檢查來源目錄中的差異檔案
	for _, file := range result.OnlyInSource {
		if !shouldIgnore(file, ignorePatterns) {
			return false
		}
	}

	// 檢查目標目錄中的差異檔案
	for _, file := range result.OnlyInTarget {
		if !shouldIgnore(file, ignorePatterns) {
			return false
		}
	}

	return true
}

// shouldIgnore 檢查檔案是否應該被忽略
func shouldIgnore(file string, patterns []string) bool {
	for _, pattern := range patterns {
		// 支援完整檔名比對
		if file == pattern {
			return true
		}
		// 支援副檔名比對
		if strings.HasPrefix(pattern, "*.") && strings.HasSuffix(file, pattern[1:]) {
			return true
		}
		// 支援前綴比對
		if strings.HasSuffix(pattern, "*") && strings.HasPrefix(file, pattern[:len(pattern)-1]) {
			return true
		}
		// 支援後綴比對
		if strings.HasPrefix(pattern, "*") && strings.HasSuffix(file, pattern[1:]) {
			return true
		}
	}
	return false
}

// getFileList 取得目錄中的所有檔案
func getFileList(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			// 取得相對於起始目錄的路徑
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}
		return nil
	})
	return files, err
}

// PrintResult 印出比對結果
func PrintResult(result *CompareResult) {
	fmt.Println("檔案差異：")
	fmt.Println("----------------------------------------")

	if len(result.OnlyInSource) > 0 {
		fmt.Println("\n只在來源目錄存在的檔案：")
		for _, file := range result.OnlyInSource {
			fmt.Printf("  %s\n", file)
		}
	}

	if len(result.OnlyInTarget) > 0 {
		fmt.Println("\n只在目標目錄存在的檔案：")
		for _, file := range result.OnlyInTarget {
			fmt.Printf("  %s\n", file)
		}
	}

	if len(result.OnlyInSource) == 0 && len(result.OnlyInTarget) == 0 {
		fmt.Println("兩個目錄的檔案完全相同")
	}
}
