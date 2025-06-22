package geocoding

import (
	"os"
	"runtime/pprof"
	"testing"
)

func TestGeocoderLocationMapping(t *testing.T) {
	testJSONPath := "/Users/kimi/go/src/photo-sorter/geodata/states.geojson"

	geocoder, err := NewGeocoder(GeoTypeJson, map[string]interface{}{
		"db_path": testJSONPath,
	})
	if err != nil {
		t.Fatalf("建立地理編碼器失敗: %v", err)
	}

	tests := []struct {
		name     string
		lat      float64
		lon      float64
		expected string
	}{
		{
			name:     "台北",
			lat:      25.0330,
			lon:      121.5654,
			expected: "New Taipei",
		},
		{
			name:     "澎湖",
			lat:      23.5494003,
			lon:      119.5890471,
			expected: "Penghu",
		},
		{
			name:     "日本東京",
			lat:      35.6895,
			lon:      139.6917,
			expected: "Tokyo",
		},
		{
			name:     "美國紐約",
			lat:      40.7128,
			lon:      -74.0060,
			expected: "New York",
		},
		{
			name:     "英國倫敦",
			lat:      51.5074,
			lon:      -0.1278,
			expected: "Westminster",
		},
		{
			name:     "法國巴黎",
			lat:      48.8566,
			lon:      2.3522,
			expected: "Paris",
		},
		{
			name:     "德國柏林",
			lat:      52.5200,
			lon:      13.4050,
			expected: "Brandenburg",
		},
		{
			name:     "中國北京",
			lat:      39.9384151,
			lon:      116.0671435,
			expected: "Beijing",
		},
		{
			name:     "韓國首爾",
			lat:      37.5665,
			lon:      126.9780,
			expected: "Gyeonggi",
		},
		{
			name:     "澳洲雪梨",
			lat:      -33.8688,
			lon:      151.2093,
			expected: "New South Wales",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location, err := geocoder.GetLocationFromGPS(tt.lat, tt.lon)
			if err != nil {
				t.Errorf("取得位置失敗: %v", err)
				return
			}

			if location.City != tt.expected {
				t.Errorf("位置不匹配，期望 %s，得到 %s", tt.expected, location.City)
			}
		})
	}
}

func TestGeocoderLocationMappingSpatialite(t *testing.T) {
	testJSONPath := "/Users/kimi/go/src/photo-sorter/geodata/states.sqlite"

	geocoder, err := NewGeocoder(GeoTypeSpatialite, map[string]interface{}{
		"db_path": testJSONPath,
	})
	if err != nil {
		t.Fatalf("建立地理編碼器失敗: %v", err)
	}

	tests := []struct {
		name     string
		lat      float64
		lon      float64
		expected string
	}{
		{
			name:     "台北",
			lat:      25.0330,
			lon:      121.5654,
			expected: "Taipei",
		},
		{
			name:     "澎湖",
			lat:      23.5494003,
			lon:      119.5890471,
			expected: "Penghu",
		},
		{
			name:     "日本東京",
			lat:      35.6895,
			lon:      139.6917,
			expected: "Tokyo",
		},
		{
			name:     "美國紐約",
			lat:      40.7128,
			lon:      -74.0060,
			expected: "New York",
		},
		{
			name:     "英國倫敦",
			lat:      51.5074,
			lon:      -0.1278,
			expected: "Westminster",
		},
		{
			name:     "法國巴黎",
			lat:      48.8566,
			lon:      2.3522,
			expected: "Paris",
		},
		{
			name:     "德國柏林",
			lat:      52.5200,
			lon:      13.4050,
			expected: "Berlin",
		},
		{
			name:     "中國北京",
			lat:      39.9384151,
			lon:      116.0671435,
			expected: "Beijing",
		},
		{
			name:     "韓國首爾",
			lat:      37.5665,
			lon:      126.9780,
			expected: "Seoul",
		},
		{
			name:     "澳洲雪梨",
			lat:      -33.8688,
			lon:      151.2093,
			expected: "New South Wales",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location, err := geocoder.GetLocationFromGPS(tt.lat, tt.lon)
			if err != nil {
				t.Errorf("取得位置失敗: %v", err)
				return
			}

			if location.City != tt.expected {
				t.Errorf("位置不匹配，期望 %s，得到 %s", tt.expected, location.City)
			}
		})
	}
}
func BenchmarkGetLocationFromGPS(b *testing.B) {
	testJSONPath := "/Users/kimi/go/src/photo-sorter/geodata/states.geojson"

	geocoder, err := NewGeocoder(GeoTypeJson, map[string]interface{}{
		"db_path": testJSONPath,
	})
	if err != nil {
		b.Fatalf("建立地理編碼器失敗: %v", err)
	}

	// 測試不同位置的效能
	testCases := []struct {
		name string
		lat  float64
		lon  float64
	}{
		{"台北", 25.0330, 121.5654},
		{"澎湖", 23.5494003, 119.5890471},
		{"東京", 35.6895, 139.6917},
		{"紐約", 40.7128, -74.0060},
		{"倫敦", 51.5074, -0.1278},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := geocoder.GetLocationFromGPS(tc.lat, tc.lon)
				if err != nil {
					b.Fatalf("取得位置失敗: %v", err)
				}
			}
		})
	}
}

func BenchmarkGetLocationFromGPSWithPprof(b *testing.B) {
	// 建立 CPU profile
	cpuFile, err := os.Create("cpu.prof")
	if err != nil {
		b.Fatalf("建立 CPU profile 失敗: %v", err)
	}
	defer cpuFile.Close()

	// 開始 CPU profiling
	if err := pprof.StartCPUProfile(cpuFile); err != nil {
		b.Fatalf("啟動 CPU profile 失敗: %v", err)
	}
	defer pprof.StopCPUProfile()

	// 建立記憶體 profile
	memFile, err := os.Create("mem.prof")
	if err != nil {
		b.Fatalf("建立記憶體 profile 失敗: %v", err)
	}
	defer memFile.Close()
	defer pprof.WriteHeapProfile(memFile)

	testJSONPath := "/Users/kimi/go/src/photo-sorter/geodata/states.geojson"

	geocoder, err := NewGeocoder(GeoTypeJson, map[string]interface{}{
		"db_path": testJSONPath,
	})
	if err != nil {
		b.Fatalf("建立地理編碼器失敗: %v", err)
	}

	// 測試不同位置的效能
	testCases := []struct {
		name string
		lat  float64
		lon  float64
	}{
		{"台北", 25.0330, 121.5654},
		{"澎湖", 23.5494003, 119.5890471},
		{"東京", 35.6895, 139.6917},
		{"紐約", 40.7128, -74.0060},
		{"倫敦", 51.5074, -0.1278},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := geocoder.GetLocationFromGPS(tc.lat, tc.lon)
				if err != nil {
					b.Fatalf("取得位置失敗: %v", err)
				}
			}
		})
	}
}

func TestSpatialiteGeocoder(t *testing.T) {
	// 測試 Spatialite 地理編碼器
	options := map[string]interface{}{
		"db_path": "geodata/states.sqlite",
	}

	geocoder, err := NewGeocoder(GeoTypeSpatialite, options)
	if err != nil {
		t.Skipf("無法創建 Spatialite 地理編碼器，跳過測試: %v", err)
	}

	// 測試台北的座標
	lat := 25.0330
	lon := 121.5654

	location, err := geocoder.GetLocationFromGPS(lat, lon)
	if err != nil {
		t.Logf("無法找到位置，這可能是正常的: %v", err)
		return
	}

	if location != nil {
		t.Logf("找到位置: Country=%s, City=%s", location.Country, location.City)
	}
}

func BenchmarkGeocoderComparison(b *testing.B) {
	// 測試數據路徑
	jsonPath := "/Users/kimi/go/src/photo-sorter/geodata/states.geojson"
	sqlitePath := "/Users/kimi/go/src/photo-sorter/geodata/states.sqlite"

	// 創建 JSON 地理編碼器
	jsonGeocoder, err := NewGeocoder(GeoTypeJson, map[string]interface{}{
		"db_path": jsonPath,
	})
	if err != nil {
		b.Fatalf("建立 JSON 地理編碼器失敗: %v", err)
	}

	// 創建 Spatialite 地理編碼器
	sqliteGeocoder, err := NewGeocoder(GeoTypeSpatialite, map[string]interface{}{
		"db_path": sqlitePath,
	})
	if err != nil {
		b.Fatalf("建立 Spatialite 地理編碼器失敗: %v", err)
	}

	// 測試不同位置的效能
	testCases := []struct {
		name string
		lat  float64
		lon  float64
	}{
		{"台北", 25.0330, 121.5654},
		{"澎湖", 23.5494003, 119.5890471},
		{"東京", 35.6895, 139.6917},
		{"紐約", 40.7128, -74.0060},
		{"倫敦", 51.5074, -0.1278},
		{"巴黎", 48.8566, 2.3522},
		{"柏林", 52.5200, 13.4050},
		{"北京", 39.9384151, 116.0671435},
		{"首爾", 37.5665, 126.9780},
		{"雪梨", -33.8688, 151.2093},
	}

	// 測試 JSON 地理編碼器
	b.Run("JSON_Geocoder", func(b *testing.B) {
		for _, tc := range testCases {
			b.Run(tc.name, func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_, err := jsonGeocoder.GetLocationFromGPS(tc.lat, tc.lon)
					if err != nil {
						b.Fatalf("JSON 地理編碼器取得位置失敗: %v", err)
					}
				}
			})
		}
	})

	// 測試 Spatialite 地理編碼器
	b.Run("Spatialite_Geocoder", func(b *testing.B) {
		for _, tc := range testCases {
			b.Run(tc.name, func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_, err := sqliteGeocoder.GetLocationFromGPS(tc.lat, tc.lon)
					if err != nil {
						b.Fatalf("Spatialite 地理編碼器取得位置失敗: %v", err)
					}
				}
			})
		}
	})
}

func BenchmarkGeocoderComparisonSingleLocation(b *testing.B) {
	// 測試數據路徑
	jsonPath := "/Users/kimi/go/src/photo-sorter/geodata/states.geojson"
	sqlitePath := "/Users/kimi/go/src/photo-sorter/geodata/states.sqlite"

	// 創建 JSON 地理編碼器
	jsonGeocoder, err := NewGeocoder(GeoTypeJson, map[string]interface{}{
		"db_path": jsonPath,
	})
	if err != nil {
		b.Fatalf("建立 JSON 地理編碼器失敗: %v", err)
	}

	// 創建 Spatialite 地理編碼器
	sqliteGeocoder, err := NewGeocoder(GeoTypeSpatialite, map[string]interface{}{
		"db_path": sqlitePath,
	})
	if err != nil {
		b.Fatalf("建立 Spatialite 地理編碼器失敗: %v", err)
	}

	// 使用台北座標進行測試
	lat := 25.0330
	lon := 121.5654

	// 測試 JSON 地理編碼器
	b.Run("JSON_Geocoder_Taipei", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := jsonGeocoder.GetLocationFromGPS(lat, lon)
			if err != nil {
				b.Fatalf("JSON 地理編碼器取得位置失敗: %v", err)
			}
		}
	})

	// 測試 Spatialite 地理編碼器
	b.Run("Spatialite_Geocoder_Taipei", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := sqliteGeocoder.GetLocationFromGPS(lat, lon)
			if err != nil {
				b.Fatalf("Spatialite 地理編碼器取得位置失敗: %v", err)
			}
		}
	})
}

func BenchmarkGeocoderComparisonWithMemory(b *testing.B) {
	// 測試數據路徑
	jsonPath := "/Users/kimi/go/src/photo-sorter/geodata/states.geojson"
	sqlitePath := "/Users/kimi/go/src/photo-sorter/geodata/states.sqlite"

	// 創建 JSON 地理編碼器
	jsonGeocoder, err := NewGeocoder(GeoTypeJson, map[string]interface{}{
		"db_path": jsonPath,
	})
	if err != nil {
		b.Fatalf("建立 JSON 地理編碼器失敗: %v", err)
	}

	// 創建 Spatialite 地理編碼器
	sqliteGeocoder, err := NewGeocoder(GeoTypeSpatialite, map[string]interface{}{
		"db_path": sqlitePath,
	})
	if err != nil {
		b.Fatalf("建立 Spatialite 地理編碼器失敗: %v", err)
	}

	// 測試不同位置的效能
	testCases := []struct {
		name string
		lat  float64
		lon  float64
	}{
		{"台北", 25.0330, 121.5654},
		{"澎湖", 23.5494003, 119.5890471},
		{"東京", 35.6895, 139.6917},
		{"紐約", 40.7128, -74.0060},
		{"倫敦", 51.5074, -0.1278},
	}

	// 測試 JSON 地理編碼器（包含記憶體統計）
	b.Run("JSON_Geocoder_WithMemory", func(b *testing.B) {
		for _, tc := range testCases {
			b.Run(tc.name, func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_, err := jsonGeocoder.GetLocationFromGPS(tc.lat, tc.lon)
					if err != nil {
						b.Fatalf("JSON 地理編碼器取得位置失敗: %v", err)
					}
				}
			})
		}
	})

	// 測試 Spatialite 地理編碼器（包含記憶體統計）
	b.Run("Spatialite_Geocoder_WithMemory", func(b *testing.B) {
		for _, tc := range testCases {
			b.Run(tc.name, func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_, err := sqliteGeocoder.GetLocationFromGPS(tc.lat, tc.lon)
					if err != nil {
						b.Fatalf("Spatialite 地理編碼器取得位置失敗: %v", err)
					}
				}
			})
		}
	})
}
