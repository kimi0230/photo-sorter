package geocoding

import (
	"errors"
)

type Geocoder interface {
	GetLocationFromGPS(lat, lon float64) (*CountryCity, error)
}

type CountryCity struct {
	Country string
	City    string
}

// GeocoderType 定義地理編碼器的類型
type GeocoderType string

const (
	// GeoStateType 使用 GeoJSON 檔案的地理編碼器
	GeoStateType GeocoderType = "geo_state"
	// 可以在這裡添加其他類型
)

// NewGeocoder 建立一個新的 Geocoder 實例
// geocoderType 指定要使用的地理編碼器類型
// options 是建立地理編碼器時需要的選項
func NewGeocoder(geocoderType GeocoderType, options map[string]interface{}) (Geocoder, error) {
	switch geocoderType {
	case GeoStateType:
		jsonPath, ok := options["json_path"].(string)
		if !ok {
			return nil, errors.New("json_path is required for GeoAlpha3JSON type")
		}
		return NewGeoState(jsonPath), nil
	default:
		return nil, errors.New("unsupported geocoder type")
	}
}
