package file

import (
	"fmt"
	"os"
	"testing"
)

func BenchmarkCopyFileDifferentSizes(b *testing.B) {
	sizes := []int{
		1 * 1024 * 1024,   // 1MB
		10 * 1024 * 1024,  // 10MB
		100 * 1024 * 1024, // 100MB
		500 * 1024 * 1024, // 500MB
	}

	srcFile := "test_source.txt"
	dstFile := "test_destination.txt"
	defer os.Remove(srcFile)
	defer os.Remove(dstFile)

	for _, size := range sizes {
		// 建立測試檔案
		data := make([]byte, size)
		if err := os.WriteFile(srcFile, data, 0644); err != nil {
			b.Fatal(err)
		}

		// 測試 CopyFile
		b.Run(fmt.Sprintf("CopyFile_%dMB", size/1024/1024), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := CopyFile(srcFile, dstFile); err != nil {
					b.Fatal(err)
				}
				os.Remove(dstFile)
			}
		})

		// 測試 CopyFileDirect
		b.Run(fmt.Sprintf("CopyFileDirect_%dMB", size/1024/1024), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := CopyFileDirect(srcFile, dstFile); err != nil {
					b.Fatal(err)
				}
				os.Remove(dstFile)
			}
		})

		// 測試 CopyFileWithBuffer
		b.Run(fmt.Sprintf("CopyFileWithBuffer_%dMB", size/1024/1024), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := CopyFileWithBuffer(srcFile, dstFile); err != nil {
					b.Fatal(err)
				}
				os.Remove(dstFile)
			}
		})
	}
}
