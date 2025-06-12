package stats

import "sync"

// Stats 用於追蹤處理統計
type Stats struct {
	TotalFiles       int
	SuccessCount     int
	FailureCount     int
	UnsupportedCount int
	IgnoredCount     int
	UnsupportedExts  map[string]int
	IgnoredExts      map[string]int
	mu               sync.Mutex
}

// NewStats 建立新的統計實例
func NewStats() *Stats {
	return &Stats{
		UnsupportedExts: make(map[string]int),
		IgnoredExts:     make(map[string]int),
	}
}

// IncrementSuccess 增加成功計數
func (s *Stats) IncrementSuccess() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SuccessCount++
}

// IncrementFailure 增加失敗計數
func (s *Stats) IncrementFailure() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.FailureCount++
}

// IncrementUnsupportedExt 增加不支援的檔案格式計數
func (s *Stats) IncrementUnsupportedExt(ext string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.UnsupportedExts[ext]++
	s.UnsupportedCount++
}

// IncrementIgnoredExt 增加忽略的檔案格式計數
func (s *Stats) IncrementIgnoredExt(ext string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.IgnoredExts[ext]++
	s.IgnoredCount++
}

// SetTotalFiles 設定總檔案數
func (s *Stats) SetTotalFiles(total int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalFiles = total
}

// GetStats 取得統計資訊
func (s *Stats) GetStats() Stats {
	s.mu.Lock()
	defer s.mu.Unlock()
	return *s
}
