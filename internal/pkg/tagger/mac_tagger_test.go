package tagger

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestMacOSTagger_AddTag(t *testing.T) {
	// 建立測試用的實體檔案
	testDir := "testdata"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("建立測試目錄失敗: %v", err)
	}

	testFile := filepath.Join(testDir, "test_image.jpg")
	file, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("建立測試檔案失敗: %v", err)
	}
	defer func() {
		file.Close()
		os.RemoveAll(testDir)
	}()

	// 寫入一些測試資料
	_, err = file.Write([]byte("test image content"))
	if err != nil {
		t.Fatalf("寫入測試資料失敗: %v", err)
	}

	tagger := NewMacOSTagger()

	// 測試添加標籤
	err = tagger.AddTag(testFile, "TestTag")
	if err != nil {
		t.Errorf("添加標籤失敗: %v", err)
	}

	// 列出並驗證標籤
	tags, err := tagger.ListTags(testFile)
	if err != nil {
		t.Errorf("列出標籤失敗: %v", err)
	}

	fmt.Printf("檔案 %s 的標籤: %v\n", testFile, tags)
}
