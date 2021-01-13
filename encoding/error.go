package encoding

import (
	"fmt"

	"github.com/go-spatial/geom"
)

// ErrUnknownGeometry is returned when a geometry type that is unknown is asked
// to be encoded
type ErrUnknownGeometry struct {
	Geom geom.Geometry
}

// Error fulfills the error interface
func (e ErrUnknownGeometry) Error() string {
	return fmt.Sprintf("unknown geometry: %T", e.Geom)
}

// ErrInvalidGeoJSON is a wrapper around a []byte that is invalid GeoJson
type ErrInvalidGeoJSON struct {
	GJSON []byte
}

// Error fulfills the error interface
func (e ErrInvalidGeoJSON) Error() string {
	return fmt.Sprintf("Invalid GeoJSON string: %T", string(e.GJSON))
}
