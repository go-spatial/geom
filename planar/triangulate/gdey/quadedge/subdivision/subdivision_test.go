package subdivision

import (
	"context"
	"fmt"
	"github.com/go-spatial/geom/cmp"
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
				t.Logf("points: %v", wkt.MustEncode(geom.MultiPoint(tc.Points)))
				if err1, ok := err.(quadedge.ErrInvalid); ok {
					for i, estr := range err1 {
						t.Logf("%v: %v", i, estr)
					}
				}
			}
			err = sd.Validate(context.Background())
			if err != nil {
				t.Logf("points: %v", wkt.MustEncode(geom.MultiPoint(tc.Points)))
				if err1, ok := err.(quadedge.ErrInvalid); ok {
					for i, estr := range err1 {
						t.Logf("%03v : %v", i, estr)
					}
				}
				t.Errorf(err.Error())
				return
			}
			idx := -1
			err = sd.WalkAllEdges(func(e *quadedge.Edge) error {
				idx++
				eln := e.AsLine()
				if idx >= len(tc.Lines) {
					return nil
				}

				t.Logf("line %v: \n\texp %v\n\tgot %v", idx, wkt.MustEncode(tc.Lines[idx]), wkt.MustEncode(eln))
				if !cmp.LineStringEqual(eln[:], tc.Lines[idx][:]) {
					t.Logf("points: %v", wkt.MustEncode(geom.MultiPoint(tc.Points)))
					t.Errorf("line %v, expected %v got %v", idx, wkt.MustEncode(tc.Lines[idx]), wkt.MustEncode(eln))
					dumpSD(t, sd)
					t.Logf("tc.Lines: %v", tc.Lines)
					return fmt.Errorf("failed")
				}
				return nil
			})
			if idx+1 != len(tc.Lines) {

				dumpSD(t, sd)
				if err != nil {
					t.Logf(err.Error())
				}
				t.Errorf("lines, expected %v got %v", len(tc.Lines), idx)
				return
			}
		}
	}

	tests := []tcase{
		{
			Desc: "one point",
			Points: [][2]float64{
				{0, 0},
			},
			Lines: must.ParseMultilines([]byte(
				`MULTILINESTRING ((-10 -10,0 0),(0 0,0 10),(0 10,-10 -10),(0 10,10 -10),(10 -10,-10 -10),(10 -10,0 0))`,
			)),
		},
		{
			Desc: "two points",
			Points: [][2]float64{
				{0, 0},
				{0, -6},
			},
			Lines: must.ParseMultilines([]byte(
				`MULTILINESTRING ((-10 -76,0 -6),(0 -6,0 0),(0 0,-10 -76),(0 0,0 130),(0 130,-10 -76),(0 130,10 -76),(10 -76,-10 -76),(10 -76,0 -6),(10 -76,0 0))`,
			)),
		},
		{
			Desc: "three points",
			Points: [][2]float64{
				{0, 0},
				{0, -6},
				{-6, 6},
			},
			Lines: must.ParseMultilines([]byte(
				`MULTILINESTRING ((-46 -136,-6 6),(-6 6,-3 256),(-3 256,-46 -136),(-3 256,40 -136),(40 -136,-46 -136),(40 -136,0 -6),(0 -6,-46 -136),(0 -6,-6 6),(0 -6,0 0),(0 0,-6 6),(0 0,-3 256),(0 0,40 -136))`,
			)),
		},
		{
			Desc: "four points",
			Points: [][2]float64{
				{0, 0},
				{0, -6},
				{-6, 6},
				{6, 6},
			},
			Lines: must.ParseMultilines([]byte(
				`MULTILINESTRING ((76 -136,6 6),(6 6,0 -6),(0 -6,76 -136),(0 -6,-76 -136),(-76 -136,76 -136),(-76 -136,0 256),(0 256,76 -136),(0 256,6 6),(0 256,-6 6),(-6 6,6 6),(-6 6,0 0),(0 0,6 6),(0 0,0 -6),(-6
 6,0 -6),(-6 6,-76 -136))`,
			)),
		},
		{
			Desc: "colinear folinear",
			Points: [][2]float64{
				{30, 4},
				{20, 4},
				{20, 2},
				{20, 6},
				{10, 6},
				{15, 4},
				{17, 4},
				{19, 3},
			},
			Lines: must.ReadMultilines("testdata/colinear_folinear.lines"),
		},
		{
			Desc:   "trunc something wrong with Florida",
			Points: must.ReadPoints("testdata/florida_trucated.points"),
			Lines:  must.ReadMultilines("testdata/florida_trucated_expected.lines"),
		},
		{
			Desc:   "florida_truncated_one",
			Points: must.ReadPoints("testdata/florida_trucated_2.points"),
			Lines:  must.ReadMultilines("testdata/florida_trucated_expected.lines"),
		},
		{
			Desc:   "something wrong with Florida",
			Points: must.ReadPoints("testdata/florida.points"),
			Lines:  must.ReadMultilines("testdata/florida_expected.lines"),
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
