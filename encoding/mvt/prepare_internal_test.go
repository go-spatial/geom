package mvt

import (
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
)

func TestPrepareptIntegerTruncation(t *testing.T) {

	type tcase struct {
		in   geom.Point
		out  geom.Point
		tile geom.Extent
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			got := preparept(tc.in, &tc.tile, float64(DefaultExtent))

			if !cmp.PointEqual(tc.out, got) {
				t.Errorf("expected %v got %v", tc.out, got)
			}

			if got[0] != float64(int64(got[0])) || got[1] != float64(int64(got[1])) {
				t.Errorf("result should be integer values, got %v", got)
			}
		}
	}

	tests := map[string]tcase{
		"exact coordinate": {
			in:   geom.Point{500, 500},
			out:  geom.Point{2048, 2048},
			tile: geom.Extent{0, 0, 1000, 1000},
		},
		"fractional coordinate should truncate": {
			in:   geom.Point{500.7, 500.3},
			out:  geom.Point{2050, 2046},
			tile: geom.Extent{0, 0, 1000, 1000},
		},
		"very small fraction should truncate to same integer": {
			in:   geom.Point{500.0001, 500.0001},
			out:  geom.Point{2048, 2047},
			tile: geom.Extent{0, 0, 1000, 1000},
		},
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}

func TestPrepareLinestring(t *testing.T) {

	type tcase struct {
		in   geom.LineString
		out  geom.LineString
		tile geom.Extent
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			got := preparelinestr(tc.in, &tc.tile, float64(DefaultExtent))

			if len(got) != len(tc.out) {
				t.Errorf("expected %v got %v", tc.out, got)
			}

			for i := range got {
				if !cmp.PointEqual(tc.out[i], got[i]) {
					t.Errorf("expected (%d) %v got %v", i, tc.out, got)
				}
			}
		}
	}

	tests := map[string]tcase{
		"duplicate pt simple line": {
			in:   geom.LineString{{9.0, 4090.0}, {9.0, 4090.0}},
			out:  geom.LineString{},
			tile: geom.Extent{0.0, 0.0, 4096.0, 4096.0},
		},
		"triplicate pt simple line": {
			in:   geom.LineString{{9.0, 4090.0}, {9.0, 4090.0}, {9.0, 4090.0}},
			out:  geom.LineString{},
			tile: geom.Extent{0.0, 0.0, 4096.0, 4096.0},
		},
		"triplicate pt simple line1": {
			in:   geom.LineString{{9.0, 4090.0}, {9.0, 4090.0}, {9.0, 4090.0}, {11.0, 4091.0}},
			out:  geom.LineString{{9.0, 6.0}, {11.0, 5.0}},
			tile: geom.Extent{0.0, 0.0, 4096.0, 4096.0},
		},
		"simple line": {
			in:   geom.LineString{{9.0, 4090.0}, {11.0, 4091.0}},
			out:  geom.LineString{{9.0, 6.0}, {11.0, 5.0}},
			tile: geom.Extent{0.0, 0.0, 4096.0, 4096.0},
		},
		"edge line": {
			in:   geom.LineString{{0.0, 0.0}, {4096.0, 20.0}},
			out:  geom.LineString{{0.0, 4096.0}, {4096.0, 4076.0}},
			tile: geom.Extent{0.0, 0.0, 4096.0, 4096.0},
		},
		"simple line 3pt": {
			in:   geom.LineString{{9.0, 4090.0}, {11.0, 4090.0}, {11.0, 4076.0}},
			out:  geom.LineString{{9.0, 6.0}, {11.0, 6.0}, {11.0, 20.0}},
			tile: geom.Extent{0.0, 0.0, 4096.0, 4096.0},
		},
		"scale": {
			in:   geom.LineString{{100.0, 100.0}, {300.0, 300.0}},
			out:  geom.LineString{{1024.0, 3072.0}, {3072.0, 1024.0}},
			tile: geom.Extent{0.0, 0.0, 400.0, 400.0},
		},
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}

func TestPreparePolygonClosingPoint(t *testing.T) {

	type tcase struct {
		in             geom.Polygon
		tile           geom.Extent
		expectedRings  int
		expectedPoints int
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			got := preparePolygon(tc.in, &tc.tile, float64(DefaultExtent))

			if len(got) != tc.expectedRings {
				t.Errorf("expected %d rings, got %d", tc.expectedRings, len(got))
			}

			if len(got) > 0 && len(got[0]) != tc.expectedPoints {
				t.Errorf("expected %d points in first ring, got %d", tc.expectedPoints, len(got[0]))
			}

			if len(got) > 0 && len(got[0]) > 0 {
				first := got[0][0]
				last := got[0][len(got[0])-1]
				if first[0] == last[0] && first[1] == last[1] {
					t.Errorf("polygon ring should not have duplicate closing point: first=%v, last=%v", first, last)
				}
			}
		}
	}

	tests := map[string]tcase{
		"simple closed polygon with duplicate closing point": {
			in: geom.Polygon{
				{
					{100, 100},
					{200, 100},
					{200, 200},
					{100, 200},
					{100, 100},
				},
			},
			tile:           geom.Extent{0, 0, 1000, 1000},
			expectedRings:  1,
			expectedPoints: 4,
		},
		"polygon with very close but not identical closing point": {
			in: geom.Polygon{
				{
					{100.0, 100.0},
					{200.0, 100.0},
					{200.0, 200.0},
					{100.0, 200.0},
					{100.0001, 100.0001},
				},
			},
			tile:           geom.Extent{0, 0, 1000, 1000},
			expectedRings:  1,
			expectedPoints: 4,
		},
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}

func TestPreparePolygonWithHoles(t *testing.T) {

	type tcase struct {
		in               geom.Polygon
		tile             geom.Extent
		expectedRings    int
		minPointsPerRing int
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			got := preparePolygon(tc.in, &tc.tile, float64(DefaultExtent))

			if len(got) != tc.expectedRings {
				t.Errorf("expected %d rings, got %d", tc.expectedRings, len(got))
			}

			for i, ring := range got {
				if len(ring) < tc.minPointsPerRing {
					t.Errorf("ring %d has too few points: %d", i, len(ring))
					continue
				}
				first := ring[0]
				last := ring[len(ring)-1]
				if first[0] == last[0] && first[1] == last[1] {
					t.Errorf("ring %d should not have duplicate closing point", i)
				}
			}
		}
	}

	tests := map[string]tcase{
		"polygon with hole": {
			in: geom.Polygon{
				{
					{100, 100},
					{900, 100},
					{900, 900},
					{100, 900},
					{100, 100},
				},
				{
					{300, 300},
					{700, 300},
					{700, 700},
					{300, 700},
					{300, 300},
				},
			},
			tile:             geom.Extent{0, 0, 1000, 1000},
			expectedRings:    2,
			minPointsPerRing: 3,
		},
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}