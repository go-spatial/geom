package geom

import (
	"reflect"
	"testing"
)

func TestMultiLineString(t *testing.T) {
	var (
		mls, mls2 MultiLineStringSetter
	)
	mls = &MultiLineString{{{10, 20}, {30, 40}}, {{-10, -5}, {15, 20}}}
	mls2 = &MultiLineString{{{15, 20}, {35, 40}}, {{-15, -5}, {20, 20}}}
	mls.SetLineStrings(mls2.LineStrings())
	if !reflect.DeepEqual(mls, mls2) {
		t.Errorf("Output (%+v) does not match expected (%+v).", mls, mls2)
	}
	mls.Points()
}
