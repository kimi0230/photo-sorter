package tagger

import (
	"fmt"
	"runtime"
)

// Tagger 處理檔案標籤的介面
type Tagger interface {
	AddTag(path string, tag string) error
	RemoveTag(path string, tag string) error
	ListTags(path string) ([]string, error)
}

// NewTagger 建立新的 Tagger 實例
// 根據作業系統返回對應的實作
func NewTagger() (Tagger, error) {
	switch runtime.GOOS {
	case "darwin":
		return NewMacOSTagger(), nil
	case "linux":
		return NewLinuxTagger(), nil
	default:
		return nil, fmt.Errorf("不支援的作業系統: %s", runtime.GOOS)
	}
}
