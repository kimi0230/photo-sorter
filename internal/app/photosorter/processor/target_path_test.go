package processor

import (
	"os"
	"path/filepath"
	"testing"

	"photo-sorter/internal/pkg/config"
	"photo-sorter/internal/pkg/metadata"
)

func TestGetTargetPath(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cfg := &config.Config{
		DstDir:       tmpDir,
		DateFormat:   "2006-01",
		EnableGeoTag: false,
	}

	t.Run("uses_date_and_model", func(t *testing.T) {
		exif := &metadata.ExifData{
			CreateDate: "2024:06:01 12:34:56",
			Model:      "iPhone 15 Pro",
		}

		target, _, err := GetTargetPath("/source/IMG_0001.JPG", exif, cfg)
		if err != nil {
			t.Fatalf("GetTargetPath error: %v", err)
		}

		expectedDir := filepath.Join(tmpDir, "2024-06", "iPhone_15_Pro")
		if filepath.Dir(target) != expectedDir {
			t.Fatalf("unexpected target dir: got %s want %s", filepath.Dir(target), expectedDir)
		}
		if _, err := os.Stat(expectedDir); err != nil {
			t.Fatalf("expected target dir to exist: %v", err)
		}
	})

	t.Run("falls_back_to_unknown_and_screenshot", func(t *testing.T) {
		exif := &metadata.ExifData{
			CreateDate:  "0000:00:00 00:00:00",
			Description: "Screenshot",
		}

		target, _, err := GetTargetPath("/source/IMG_0002.JPG", exif, cfg)
		if err != nil {
			t.Fatalf("GetTargetPath error: %v", err)
		}

		expectedDir := filepath.Join(tmpDir, "unknown_date", "Screenshot")
		if filepath.Dir(target) != expectedDir {
			t.Fatalf("unexpected target dir: got %s want %s", filepath.Dir(target), expectedDir)
		}
	})
}
