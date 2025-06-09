package geocoding

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type GeoState struct {
	Alpha3   string `json:"id"`
	Name     string `json:"name"`
	Country  string `json:"country"`
	jsonPath string
	// 快取 GeoJSON 資料
	collection *GeoJSONCollection
}

// NewGeoState 建立一個新的 GeoState 實例
func NewGeoState(jsonPath string) (*GeoState, error) {
	gs := &GeoState{
		jsonPath: jsonPath,
	}

	// 初始化時載入資料
	if err := gs.loadGeoJSON(); err != nil {
		return nil, fmt.Errorf("載入 GeoJSON 失敗: %w", err)
	}

	return gs, nil
}

// loadGeoJSON 載入 GeoJSON 資料
func (g *GeoState) loadGeoJSON() error {
	jsonFile, err := os.Open(g.jsonPath)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	g.collection = &GeoJSONCollection{}
	if err := json.Unmarshal(byteValue, g.collection); err != nil {
		return err
	}

	return nil
}

type GeoJSONFeature struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	Properties struct {
		Name   string `json:"name"`
		Admin  string `json:"admin"`
		Adm0A3 string `json:"adm0_a3"`
	} `json:"properties"`
	Geometry struct {
		Type        string          `json:"type"`
		Coordinates json.RawMessage `json:"coordinates"`
	} `json:"geometry"`
}

type GeoJSONCollection struct {
	Type     string           `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}

func (g *GeoState) GetLocationFromGPS(lat, lon float64) (*CountryCity, error) {
	if g.collection == nil {
		return nil, errors.New("GeoJSON 資料未載入")
	}

	// 檢查每個多邊形是否包含給定的座標
	for _, feature := range g.collection.Features {
		switch feature.Geometry.Type {
		case "Polygon":
			var coordinates [][][]float64
			if err := json.Unmarshal(feature.Geometry.Coordinates, &coordinates); err != nil {
				continue
			}
			if len(coordinates) > 0 && len(coordinates[0]) > 0 {
				if isPointInPolygon(lat, lon, coordinates[0]) {
					countryCity := &CountryCity{
						Country: feature.Properties.Adm0A3,
						City:    feature.Properties.Name,
					}
					return countryCity, nil
				}
			}
		case "MultiPolygon":
			var coordinates [][][][]float64
			if err := json.Unmarshal(feature.Geometry.Coordinates, &coordinates); err != nil {
				continue
			}
			// 檢查每個多邊形
			for _, polygon := range coordinates {
				if len(polygon) > 0 && len(polygon[0]) > 0 {
					if isPointInPolygon(lat, lon, polygon[0]) {
						countryCity := &CountryCity{
							Country: feature.Properties.Adm0A3,
							City:    feature.Properties.Name,
						}
						return countryCity, nil
					}
				}
			}
		}
	}

	return nil, errors.New("location not found")
}

// isPointInPolygon 使用射線法判斷點是否在多邊形內
// GeoJSON 中的座標順序是 [經度, 緯度]
func isPointInPolygon(lat, lon float64, polygon [][]float64) bool {
	inside := false
	j := len(polygon) - 1

	for i := 0; i < len(polygon); i++ {
		if (polygon[i][1] > lat) != (polygon[j][1] > lat) &&
			lon < (polygon[j][0]-polygon[i][0])*(lat-polygon[i][1])/(polygon[j][1]-polygon[i][1])+polygon[i][0] {
			inside = !inside
		}
		j = i
	}

	return inside
}

// FormatCity 將城市名稱中的空白替換為底線
func (c *CountryCity) FormatCity() string {
	return strings.ReplaceAll(c.City, " ", "_")
}
