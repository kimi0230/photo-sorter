package tagger

// LinuxTagger Linux 專用的標籤實作
type LinuxTagger struct{}

// NewLinuxTagger 建立新的 LinuxTagger 實例
func NewLinuxTagger() *LinuxTagger {
	return &LinuxTagger{}
}

// AddTag 為檔案添加標籤
func (t *LinuxTagger) AddTag(path string, tag string) error {
	// TODO: 實作 Linux 的標籤功能
	return nil
}

// RemoveTag 移除檔案的標籤
func (t *LinuxTagger) RemoveTag(path string, tag string) error {
	// TODO: 實作 Linux 的標籤功能
	return nil
}

// ListTags 列出檔案的所有標籤
func (t *LinuxTagger) ListTags(path string) ([]string, error) {
	// TODO: 實作 Linux 的標籤功能
	return nil, nil
}
