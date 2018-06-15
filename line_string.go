package geom

import "errors"

var ErrNilLineString = errors.New("geom: nil LineString")

// LineString is a basic line type which is made up of two or more points that don't interacted.
type LineString [][2]float64

/* We need to change this to a WalkPoints function.
// Points returns a slice of XY values
func (ls LineString) Points() [][2]float64 {
	return ls
}
*/

/*
Returns true if the first and last vertices are the same
*/
func (ls LineString) IsRing() bool {
	last := len(ls) - 1
	if len(ls) > 1 && ls[0][0] == ls[last][0] && ls[0][1] == ls[last][1] {
		return true
	}
	return false
}

// Vertexes returns a slice of XY values
func (ls LineString) Verticies() [][2]float64 {
	return ls
}

// SetVertexes modifies the array of 2D coordinates
func (ls *LineString) SetVerticies(input [][2]float64) (err error) {
	if ls == nil {
		return ErrNilLineString
	}

	*ls = append((*ls)[:0], input...)
	return
}
