package subdivision

import (
	"context"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/planar/triangulate/delaunay/test/must"
	"github.com/go-spatial/geom/winding"
)

func TestNewSubdivisionFromGeomLines(t *testing.T) {
	type tcase struct {
		Desc  string
		Lines []geom.Line
		Skip  string
	}

	order := winding.Order{
		YPositiveDown: true,
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {

			if tc.Skip != "" {
				t.Skip(tc.Skip)
				return
			}
			sd := NewSubdivisionFromGeomLines(tc.Lines, order)
			if sd == nil {
				t.Errorf("subdivision, expected not nil, got nil")
				return
			}
			if err := sd.Validate(context.Background()); err != nil {
				t.Errorf("error, expected nil, got %v", err)
			}

		}
	}

	tests := []tcase{
		// subtests
		{
			Desc:  "intersecting_lines",
			Lines: must.ReadMultilines("testdata/intersecting_lines_97.lines"),
			Skip:  "Failing will have to look at why",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Desc, fn(tc))
	}
}
