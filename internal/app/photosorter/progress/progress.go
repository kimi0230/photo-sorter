package progress

import "sync"

// Progress 用於追蹤處理進度
type Progress struct {
	mu             sync.RWMutex
	processedFiles int
	totalFiles     int
}

// NewProgress 建立新的進度追蹤實例
func NewProgress() *Progress {
	return &Progress{}
}

// Update 更新已處理檔案數
func (p *Progress) Update() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.processedFiles++
}

// SetTotal 設定總檔案數
func (p *Progress) SetTotal(total int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.totalFiles = total
}

// GetStatus 取得處理狀態
func (p *Progress) GetStatus() (processed, total int) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.processedFiles, p.totalFiles
}
