package geocoding

import (
	"testing"
)

func TestNewGeocoder(t *testing.T) {

	testJSONPath := "./countries.geo.json"

	tests := []struct {
		name         string
		geocoderType GeocoderType
		options      map[string]interface{}
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "成功建立 GeoAlpha3JSON 地理編碼器",
			geocoderType: GeoStateType,
			options: map[string]interface{}{
				"json_path": testJSONPath,
			},
			wantErr: false,
		},
		{
			name:         "缺少 json_path 選項",
			geocoderType: GeoStateType,
			options:      map[string]interface{}{},
			wantErr:      true,
			errMsg:       "json_path is required for GeoAlpha3JSON type",
		},
		{
			name:         "不支援的地理編碼器類型",
			geocoderType: "unsupported_type",
			options:      map[string]interface{}{},
			wantErr:      true,
			errMsg:       "unsupported geocoder type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			geocoder, err := NewGeocoder(tt.geocoderType, tt.options)

			if tt.wantErr {
				if err == nil {
					t.Error("期望錯誤但沒有得到錯誤")
				} else if err.Error() != tt.errMsg {
					t.Errorf("錯誤訊息不匹配，期望 %q，得到 %q", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("不期望的錯誤: %v", err)
				return
			}

			if geocoder == nil {
				t.Error("地理編碼器不應該為 nil")
			}

			// 測試地理編碼器功能
			if tt.geocoderType == GeoStateType {
				location, err := geocoder.GetLocationFromGPS(23.5, 121.0)
				if err != nil {
					t.Errorf("GetLocationFromGPS 失敗: %v", err)
				}
				if location != "Taiwan" {
					t.Errorf("期望位置為 Taiwan，得到 %s", location)
				}
			}
		})
	}
}

func TestGeocoderLocationMapping(t *testing.T) {
	testJSONPath := "./geodata/states.geojson"

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

			if location != tt.expected {
				t.Errorf("位置不匹配，期望 %s，得到 %s", tt.expected, location)
			}
		})
	}
}
