package slippy

import (
	"testing"
	"math/rand"
	"github.com/go-spatial/proj"
	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
)

func TestTileGridSize(t *testing.T) {
	type tcase struct {
		srid          uint
		zoom          uint
		expectedSizeX uint
		expectedSizeY uint
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			grid, err := NewGrid(tc.srid)
			if err != nil {
				t.Fatal(err)
			}
			if grid.SRID() != tc.srid {
				t.Fatal(err)
			}
			tile, ok := grid.Size(tc.zoom)
			if !ok {
				t.Fatal("expected ok")
			}
			if tile.X != tc.expectedSizeX {
				t.Errorf("got %v expected %v", tile.X, tc.expectedSizeX)
			}
			if tile.Y != tc.expectedSizeY {
				t.Errorf("got %v expected %v", tile.Y, tc.expectedSizeY)
			}
		}
	}

	tests := map[string]tcase{
		"4326_zoom0": {
			srid:          4326,
			zoom:          0,
			expectedSizeX: 2,
			expectedSizeY: 1,
		},
		"3857_zoom0": {
			srid:          3857,
			zoom:          0,
			expectedSizeX: 1,
			expectedSizeY: 1,
		},
		"4326_zoom15": {
			srid:          4326,
			zoom:          15,
			expectedSizeX: 65536,
			expectedSizeY: 32768,
		},
		"3857_zoom15": {
			srid:          3857,
			zoom:          15,
			expectedSizeX: 32768,
			expectedSizeY: 32768,
		},
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}

func TestTileGridContains(t *testing.T) {
	type tcase struct {
		srid     uint
		zoom     uint
		x        uint
		y        uint
		expected bool
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			grid, err := NewGrid(tc.srid)
			if err != nil {
				t.Fatal(err)
			}

			_, ok := grid.ToNative(NewTile(tc.zoom, tc.x, tc.y))

			if ok != tc.expected {
				t.Errorf("got %v expected %v", ok, tc.expected)
			}
		}
	}

	tests := map[string]tcase{
		"3857_zoom0_pass": {
			srid:     3857,
			zoom:     0,
			x:        0,
			y:        0,
			expected: true,
		},
		"3857_zoom0_fail": {
			srid:     3857,
			zoom:     0,
			x:        2,
			y:        0,
			expected: false,
		},
		"3857_zoom15_extent": {
			srid:     3857,
			zoom:     15,
			x:        32767,
			y:        32767,
			expected: true,
		},
		"4326_zoom0_pass": {
			srid:     4326,
			zoom:     0,
			x:        2,
			y:        0,
			expected: true,
		},
		"4326_zoom0_fail": {
			srid:     4326,
			zoom:     0,
			x:        0,
			y:        2,
			expected: false,
		},
		"4326_zoom12_pass": {
			srid:     4326,
			zoom:     12,
			x:        8191,
			y:        4095,
			expected: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}

func TestFromNative(t *testing.T) {
	type tcase struct {
		point    geom.Point
		srid     uint
		zoom     uint
		expected [2]uint
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			grid, err := NewGrid(tc.srid)
			if err != nil {
				t.Fatal(err)
			}

			pt := tc.point
			if tc.srid != 4326 {
				pts, err := proj.Convert(proj.EPSGCode(tc.srid), pt[:])
				if err != nil {
					t.Fatal(err, tc.srid)
				}
				pt = geom.Point{pts[0], pts[1]}
			}

			tile, ok := grid.FromNative(tc.zoom, pt)
			if !ok {
				t.Fatal("expected ok")
			}

			if tile.X != tc.expected[0] {
				t.Errorf("got %v expected %v", tile.X, tc.expected[0])
			}

			if tile.Y != tc.expected[1] {
				t.Errorf("got %v expected %v", tile.Y, tc.expected[1])
			}
		}
	}
	
	// expected = tile column, tile row
	tests := map[string]tcase{
		"3857_z0": {
			point:    geom.Point{0.0, 0.0},
			srid:     3857,
			zoom:     0,
			expected: [2]uint{0, 0},
		},
		"3857_z0_random": {
			point:    geom.Point{0, 0 + rand.Float64() * 85.0511},
			srid:     3857,
			zoom:     0,
			expected: [2]uint{0, 0},
		},
		"3857_z10_quad1": {
			point:    geom.Point{179.99999, 85.0511},
			srid:     3857,
			zoom:     10,
			expected: [2]uint{1023, 0},
		},
		"3857_z10_quad2": {
			point:    geom.Point{-179.99999, 85.0511},
			srid:     3857,
			zoom:     10,
			expected: [2]uint{0, 0},
		},
		"3857_z10_quad3": {
			point:    geom.Point{-179.99999, -85.0511},
			srid:     3857,
			zoom:     10,
			expected: [2]uint{0, 1023},
		},
		"3857_z10_quad4": {
			point:    geom.Point{179.99999, -85.0511},
			srid:     3857,
			zoom:     10,
			expected: [2]uint{1023, 1023},
		},
		"4326_z0_quad1": {
			point:    geom.Point{0.0, 0.0},
			srid:     4326,
			zoom:     0,
			expected: [2]uint{1, 0},
		},
		"4326_z0_quad2": {
			point:    geom.Point{-1.0, 0.0},
			srid:     4326,
			zoom:     0,
			expected: [2]uint{0, 0},
		},
		"4326_z10_quad1": {
			point:    geom.Point{179.99999, 89.99999},
			srid:     4326,
			zoom:     10,
			expected: [2]uint{2047, 0},
		},
		"4326_z10_quad2": {
			point:    geom.Point{-179.99999, 89.99999},
			srid:     4326,
			zoom:     10,
			expected: [2]uint{0, 0},
		},
		"4326_z10_quad3": {
			point:    geom.Point{-179.99999, -89.99999},
			srid:     4326,
			zoom:     10,
			expected: [2]uint{0, 1023},
		},
		"4326_z10_quad4": {
			point:    geom.Point{179.99999, -89.99999},
			srid:     4326,
			zoom:     10,
			expected: [2]uint{2047, 1023},
		},
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}

func TestToNative(t *testing.T) {
	type tcase struct {
		x        uint
		y        uint
		srid     uint
		zoom     uint
		expected [2]float64
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			grid, err := NewGrid(tc.srid)
			if err != nil {
				t.Fatal(err)
			}

			pt, ok := grid.ToNative(NewTile(tc.zoom, tc.x, tc.y))
			if !ok {
				t.Fatal("expected ok")
			}

			if tc.srid != 4326 {
				pts, err := proj.Inverse(proj.EPSGCode(tc.srid), pt[:])
				if err != nil {
					t.Fatal(err, tc.srid, pt)
				}
				pt = geom.Point{pts[0], pts[1]}
			}

			if !cmp.Float(pt.X(), tc.expected[0]) {
				t.Errorf("got %v expected %v", pt.X(), tc.expected[0])
			}

			if !cmp.Float(pt.Y(), tc.expected[1]) {
				t.Errorf("got %v expected %v", pt.Y(), tc.expected[1])
			}
		}
	}

	tests := map[string]tcase{
		"3857_z0": {
			x:        0,
			y:		  0,
			srid:     3857,
			zoom:     0,
			expected: [2]float64{-180.0, 85.0511},
		},
		"3857_z10_q1": {
			x:        1023,
			y:		  0,
			srid:     3857,
			zoom:     10,
			expected: [2]float64{179.6484375, 85.0511},
		},
		"3857_z10_q2": {
			x:        0,
			y:		  0,
			srid:     3857,
			zoom:     10,
			expected: [2]float64{-180.0, 85.0511},
		},
		"3857_z10_q3": {
			x:        0,
			y:		  1023,
			srid:     3857,
			zoom:     10,
			expected: [2]float64{-180.0, -85.0207},
		},
		"3857_z10_q4": {
			x:        1023,
			y:		  1023,
			srid:     3857,
			zoom:     10,
			expected: [2]float64{179.6484375, -85.0207},
		},
		"4326_z0_q1": {
			x:        1,
			y:		  0,
			srid:     4326,
			zoom:     0,
			expected: [2]float64{0, 90},
		},
		"4326_z0_q2": {
			x:        0,
			y:		  0,
			srid:     4326,
			zoom:     0,
			expected: [2]float64{-180, 90},
		},
		"4326_z10_q1": {
			x:        2047,
			y:		  0,
			srid:     4326,
			zoom:     10,
			expected: [2]float64{179.8242, 89.99999},
		},
		"4326_z10_q2": {
			x:        0,
			y:		  0,
			srid:     4326,
			zoom:     10,
			expected: [2]float64{-179.99999, 89.99999},
		},
		"4326_z10_q3": {
			x:        0,
			y:		  1023,
			srid:     4326,
			zoom:     10,
			expected: [2]float64{-179.99999, -89.8242},
		},
		"4326_z10_q4": {
			x:        2047,
			y:		  1023,
			srid:     4326,
			zoom:     10,
			expected: [2]float64{179.8242, -89.8242},
		},
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}
