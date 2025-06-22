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
	GeoTypeJson GeocoderType = "geo_json"
	// 可以在這裡添加其他類型
	GeoTypeSpatialite GeocoderType = "geo_spatialite"
)

// NewGeocoder 建立一個新的 Geocoder 實例
// geocoderType 指定要使用的地理編碼器類型
// options 是建立地理編碼器時需要的選項
func NewGeocoder(geocoderType GeocoderType, options map[string]interface{}) (Geocoder, error) {
	switch geocoderType {
	case GeoTypeJson:
		jsonPath, ok := options["db_path"].(string)
		if !ok {
			return nil, errors.New("db_path is required for GeoAlpha3JSON type")
		}
		return NewGeoStateJson(jsonPath)
	case GeoTypeSpatialite:
		dbPath, ok := options["db_path"].(string)
		if !ok {
			return nil, errors.New("db_path is required for GeoAlpha3Spatialite type")
		}
		return NewGeoSpatialite(dbPath)
	default:
		return nil, errors.New("unsupported geocoder type")
	}
}
