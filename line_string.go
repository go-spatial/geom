package geom

import "fmt"

// LineString is a basic line type which is made up of two or more points that don't interect.
type LineString [][2]float64

// Points returns a slice of XY values
func (ls *LineString) Points() [][2]float64 {
	return *ls
}

// SetPoints modifies the array of 2D coordinates
func (ls *LineString) SetPoints(input [][2]float64) (err error) {
	if ls == nil {
		return fmt.Errorf("LineString is nil")
	}

	*ls = append((*ls)[:0], input...)
	return
}
