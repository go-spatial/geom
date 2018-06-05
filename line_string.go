package geom

import (
	"errors"
)

var ErrNilLineString = errors.New("geom: nil LineString")
var ErrInvalidLineString = errors.New("geom: invalid LineString")

// LineString is a basic line type which is made up of two or more points that don't interacted.
type LineString [][2]float64

// Vertexes returns a slice of XY values
func (ls LineString) Verticies() [][2]float64 { return ls }

// SetVertexes modifies the array of 2D coordinates
func (ls *LineString) SetVerticies(input [][2]float64) (err error) {
	if ls == nil {
		return ErrNilLineString
	}

	*ls = append((*ls)[:0], input...)
	return
}

// AsSegments returns the line string as a slice of lines.
func (ls LineString) AsSegments() (segs []Line, err error) {
	switch len(ls) {
	case 0:
		return nil, nil
	case 1:
		return nil, ErrInvalidLineString
	case 2:
		return []Line{{ls[0], ls[1]}}, nil
	default:
		segs = make([]Line, len(ls)-1)
		for i := 0; i < len(ls)-1; i++ {
			segs[i] = Line{ls[i], ls[i+1]}
		}
		return segs, nil
	}
}
