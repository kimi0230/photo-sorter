package geocoding

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type GeoSpatialite struct {
	dbPath string
}

func NewGeoSpatialite(dbPath string) (*GeoSpatialite, error) {
	gs := &GeoSpatialite{
		dbPath: dbPath,
	}
	return gs, nil
}

func (gs *GeoSpatialite) GetLocationFromGPS(lat, lon float64) (*CountryCity, error) {
	pointWKT := fmt.Sprintf("POINT(%f %f)", lon, lat)

	// 建立 spatialite 查詢語句
	sql := fmt.Sprintf(`
	SELECT name_en, admin, adm0_a3 FROM ne_10m_admin_1_states_provinces
	WHERE ST_Contains(
		GEOMETRY,
		ST_PointFromText('%s', 4326)
	);
	`, pointWKT)

	// 使用 tools/bin 下的 spatialite
	// spatialitePath := filepath.Join("tools", "bin", "spatialite")
	spatialitePath := filepath.Join("spatialite")
	cmd := exec.Command(spatialitePath, gs.dbPath)

	// 傳入 SQL 並執行
	cmd.Stdin = strings.NewReader(sql)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("spatialite command error: %v, output: %s", err, out.String())
	}

	result := strings.TrimSpace(out.String())

	// 解析結果
	// 預期格式: name|admin|adm0_a3
	parts := strings.Split(result, "|")
	if len(parts) >= 3 {
		return &CountryCity{
			Country: strings.TrimSpace(parts[2]), // adm0_a3
			City:    strings.TrimSpace(parts[0]), // name
		}, nil
	}

	return nil, errors.New("location not found")
}
