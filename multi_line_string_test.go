package geom

import (
	"testing"
)

func TestMultiLineString(t *testing.T) {
	var (
		mls MultiLineStringer
	)
	mls = &MultiLineString{{{10, 20}, {30, 40}}, {{-10, -5}, {15, 20}}}
	mls.LineStrings()
	mls.Points()
}
