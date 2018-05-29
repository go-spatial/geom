package encoding

import (
	"fmt"

	"github.com/go-spatial/geom"
)

type ErrUnknownGeometry struct {
	Geom geom.Geometry
}

type ErrInvalidGeoJSON struct {
	GJSON []byte
}

func (e ErrUnknownGeometry) Error() string {
	return fmt.Sprintf("unknown geometry: %T", e.Geom)
}

func (e ErrInvalidGeoJSON) Error() string {
	return fmt.Sprintf("Invalid GeoJSON string: %T", string(e.GJSON))
}
