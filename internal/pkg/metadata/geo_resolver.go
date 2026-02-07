package metadata

import (
	"fmt"
	"strings"
	"sync"

	"photo-sorter/internal/pkg/config"
	"photo-sorter/internal/pkg/geocoding"
)

type defaultGeoResolver struct {
	geocoder geocoding.Geocoder
	cache    map[string]*geocoding.CountryCity
	mu       sync.RWMutex
}

// NewDefaultGeoResolver initializes a geo resolver with a shared geocoder instance.
func NewDefaultGeoResolver(cfg *config.Config) GeoResolver {
	var geocoder geocoding.Geocoder
	if cfg != nil && cfg.EnableGeoTag {
		var err error
		geocoder, err = geocoding.NewGeocoder(cfg.GeocoderType, map[string]interface{}{
			"db_path": cfg.GeoDBPath,
		})
		if err != nil {
			Warnf("failed to initialize geocoder: %v", err)
		}
	}
	return &defaultGeoResolver{
		geocoder: geocoder,
		cache:    make(map[string]*geocoding.CountryCity),
	}
}

func (r *defaultGeoResolver) Resolve(baseDate string, exif *ExifData, cfg *config.Config) (string, *geocoding.CountryCity, error) {
	// If geo tagging is disabled or GPS data is missing, skip geo lookup.
	if cfg == nil || !cfg.EnableGeoTag || exif.GPSLatitude == "" || exif.GPSLongitude == "" {
		return baseDate, nil, nil
	}
	if r == nil || r.geocoder == nil {
		return baseDate, nil, nil
	}

	lat, err := ParseGPSString(exif.GPSLatitude)
	if err != nil {
		return baseDate, nil, fmt.Errorf("failed to parse latitude: %v", err)
	}

	lon, err := ParseGPSString(exif.GPSLongitude)
	if err != nil {
		return baseDate, nil, fmt.Errorf("failed to parse longitude: %v", err)
	}

	if lat == 0 || lon == 0 {
		return baseDate, nil, nil
	}

	cacheKey := ""
	if cfg.EnableGeoCache {
		precision := cfg.GeoCachePrecision
		if precision <= 0 {
			precision = 5
		}
		cacheKey = fmt.Sprintf("%.*f,%.*f", precision, lat, precision, lon)
		r.mu.RLock()
		cached := r.cache[cacheKey]
		r.mu.RUnlock()
		if cached != nil {
			dateWithLocation := fmt.Sprintf("%s-%s-%s", baseDate, cached.Country, strings.ReplaceAll(cached.City, " ", "_"))
			return dateWithLocation, cached, nil
		}
	}

	countryCity, err := r.geocoder.GetLocationFromGPS(lat, lon)
	if err != nil || countryCity == nil {
		return baseDate, nil, nil
	}

	if cfg.EnableGeoCache && cacheKey != "" {
		r.mu.Lock()
		r.cache[cacheKey] = countryCity
		r.mu.Unlock()
	}

	dateWithLocation := fmt.Sprintf("%s-%s-%s", baseDate, countryCity.Country, strings.ReplaceAll(countryCity.City, " ", "_"))
	return dateWithLocation, countryCity, nil
}
