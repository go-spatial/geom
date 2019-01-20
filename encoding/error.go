package encoding

import (
	"fmt"

	"github.com/go-spatial/geom"
)

// ErrUnknownGeometry is a wrapper around a geom.Geometry that is invalid
type ErrUnknownGeometry struct {
	Geom geom.Geometry
}

// ErrInvalidGeoJSON is a wrapper around a []byte that is invalid GeoJson
type ErrInvalidGeoJSON struct {
	GJSON []byte
}

func (e ErrUnknownGeometry) Error() string {
	return fmt.Sprintf("unknown geometry: %T", e.Geom)
}

func (e ErrInvalidGeoJSON) Error() string {
	return fmt.Sprintf("Invalid GeoJSON string: %T", string(e.GJSON))
}
