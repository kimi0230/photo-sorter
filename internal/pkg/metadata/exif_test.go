package metadata

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"photo-sorter/internal/pkg/config"
)

func TestParseGPSString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		want      float64
		wantError bool
	}{
		{
			name:  "north",
			input: `22 deg 41' 58.80" N`,
			want:  22 + 41.0/60.0 + 58.8/3600.0,
		},
		{
			name:  "south",
			input: `22 deg 41' 58.80" S`,
			want:  -(22 + 41.0/60.0 + 58.8/3600.0),
		},
		{
			name:  "empty",
			input: "",
			want:  0,
		},
		{
			name:      "invalid",
			input:     "bad",
			wantError: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseGPSString(test.input)
			if test.wantError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if math.Abs(got-test.want) > 1e-6 {
				t.Fatalf("unexpected result: got %.8f want %.8f", got, test.want)
			}
		})
	}
}

func TestIsValidDate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  bool
	}{
		{"", false},
		{"0000:00:00 00:00:00", false},
		{"0000-00-00 00:00:00", false},
		{"2024:06:01 12:34:56", true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.input, func(t *testing.T) {
			t.Parallel()
			if got := isValidDate(test.input); got != test.want {
				t.Fatalf("unexpected result for %q: got %v want %v", test.input, got, test.want)
			}
		})
	}
}

func TestGetTargetPath(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cfg := &config.Config{
		DstDir:       tmpDir,
		DateFormat:   "2006-01",
		EnableGeoTag: false,
	}

	t.Run("uses_date_and_model", func(t *testing.T) {
		exif := &ExifData{
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
		exif := &ExifData{
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
