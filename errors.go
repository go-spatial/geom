package geom

import (
	"fmt"
	"github.com/gdey/errors"
)

const (
	// ErrPointsAreCoLinear is thrown when points are colinear but that is unexpected
	ErrPointsAreCoLinear = errors.String("given points are colinear")
)

// ErrUnknownGeometry represents an objects that is not a known geom geometry.
type ErrUnknownGeometry struct {
	Geom Geometry
}

func (e ErrUnknownGeometry) Error() string {
	return fmt.Sprintf("unknown geometry: %T", e.Geom)
}
