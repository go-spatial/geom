package geom

import (
	"testing"
)

func TestMultiLineString(t *testing.T) {
	var (
		mls MultiLineStringer
	)
	mls = &MultiLineString{[][2]float64{[2]float64{10, 20}, [2]float64{30, 40}},
		[][2]float64{[2]float64{-10, -5}, [2]float64{15, 20}}}
	mls.LineStrings()
	mls.Points()
}
