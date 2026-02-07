package metadata

import (
	"context"

	"photo-sorter/internal/pkg/config"
	"photo-sorter/internal/pkg/geocoding"
)

type ExifData struct {
	CreateDate      string `json:"CreateDate"`
	MediaCreateDate string `json:"MediaCreateDate"`
	DateTimeCreated string `json:"DateTimeCreated"`
	FileModifyDate  string `json:"FileModifyDate"`
	Model           string `json:"Model"`
	Encoder         string `json:"Encoder"`
	Description     string `json:"Description"`
	GPSLatitude     string `json:"GPSLatitude"`
	GPSLongitude    string `json:"GPSLongitude"`
}

//go:generate mockgen -destination=exif_mock.go -package=metadata . ExifReader
type ExifReader interface {
	GetExifData(ctx context.Context, path string) (*ExifData, error)
	Close() error
}

// Warnf is a hook for reporting non-fatal warnings to the caller.
// By default it does nothing and can be overridden by the CLI/UI layer.
var Warnf = func(format string, args ...any) {}

type DateResolver interface {
	Resolve(exif *ExifData, cfg *config.Config) (string, error)
}

type DeviceResolver interface {
	Resolve(exif *ExifData) string
}

type GeoResolver interface {
	Resolve(baseDate string, exif *ExifData, cfg *config.Config) (string, *geocoding.CountryCity, error)
}
