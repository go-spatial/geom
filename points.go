package geom

import (
	"errors"
)

// ErrNilPointS is thrown when a point is null but shouldn't be
var ErrNilPointS = errors.New("geom: nil PointS")

// Point describes a simple 2D point with SRID
type PointS struct {
	Srid
	Xy Point
}

// XYS returns the struct itself
func (p PointS) XYS() struct {
	Srid
	Xy Point
} {
	return p
}

// XY returns 2D point
func (p PointS) XY() [2]float64 {
	return p.Xy.XY()
}

// S returns the srid as uint32
func (p PointS) S() uint32 {
	return p.SRID()
}

// SetXYS sets the XY coordinates and the SRID
func (p *PointS) SetXYS(srid uint32, xy Point) (err error) {
	if p == nil {
		return ErrNilPointS
	}

	p.Srid = Srid(srid)
	p.Xy = xy
	return
}
