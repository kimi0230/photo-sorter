package geocoding

import (
	"testing"
)

func TestGeocoderLocationMapping(t *testing.T) {
	testJSONPath := "/Users/kimi/go/src/photo-sorter/geodata/states.geojson"

	geocoder, err := NewGeocoder(GeoStateType, map[string]interface{}{
		"json_path": testJSONPath,
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
