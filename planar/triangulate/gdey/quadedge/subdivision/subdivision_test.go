package subdivision

import (
	"context"
	"github.com/go-spatial/geom/encoding/wkt"
	"testing"

	"github.com/go-spatial/geom/planar/triangulate/gdey/quadedge/quadedge"
	"github.com/go-spatial/geom/planar/triangulate/gdey/quadedge/test/must"

	"github.com/go-spatial/geom"
)

func TestNewForPoints(t *testing.T) {
	type tcase struct {
		Desc   string
		Points [][2]float64
		Lines  []geom.Line
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {

			sd, err := NewForPoints(context.Background(), tc.Points)
			if err != nil {
				t.Errorf("err, expected nil got %v", err)
				t.Logf("points: %v",wkt.MustEncode(geom.MultiPoint(tc.Points)))
				if err1, ok := err.(quadedge.ErrInvalid); ok {
					for i, estr := range err1 {
						t.Logf("%v: %v", i, estr)
					}
				}
			}
			dumpSD(t, sd)

		}
	}

	tests := []tcase{
		{
			Desc: "colinear folinear",
			Points: [][2]float64{
				{3024, 4160},
				{2024, 4160},
				{2024, 2160},
				{2024, 6160},
				{1024, 6160},
				{1913, 4160},
				{2023, 4160},
				{2033, 4159},
			},
		},
		{
			Desc:   "trunc something wrong with Florida",
			Points: must.ReadPoints("testdata/florida_trucated.points"),
		},
		{
			Desc:   "something wrong with Florida",
			Points: must.ReadPoints("testdata/florida.points"),
		},
		{
			Desc:   "something wrong with north Africa",
			Points: must.ReadPoints("testdata/north_africa.points"),
		},
		{
			Desc:   "intersecting lines are generated 1",
			Points: must.ReadPoints("testdata/intersecting_lines_1.points"),
			Lines:  must.ReadMultilines("testdata/intersecting_lines_1_expected.lines"),
		},
		{
			Desc:   "counter clockwise error east of china",
			Points: must.ReadPoints("testdata/east_of_china.points"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.Desc, fn(tc))
	}
}
