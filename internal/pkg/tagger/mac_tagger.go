package tagger

import (
	"os/exec"
	"strings"
)

// MacOSTagger macOS 專用的標籤實作
type MacOSTagger struct{}

// NewMacOSTagger 建立新的 MacOSTagger 實例
func NewMacOSTagger() *MacOSTagger {
	return &MacOSTagger{}
}

// AddTag 為檔案添加標籤
func (t *MacOSTagger) AddTag(path string, tag string) error {
	cmd := exec.Command("tag", "-a", tag, path)
	return cmd.Run()
}

// RemoveTag 移除檔案的標籤
func (t *MacOSTagger) RemoveTag(path string, tag string) error {
	cmd := exec.Command("tag", "-r", tag, path)
	return cmd.Run()
}

// ListTags 列出檔案的所有標籤
func (t *MacOSTagger) ListTags(path string) ([]string, error) {
	cmd := exec.Command("tag", "-l", path)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
}
