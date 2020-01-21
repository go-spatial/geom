package slippy

import (
	"testing"
	"github.com/go-spatial/proj"
	"github.com/go-spatial/geom"
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
		expected *Tile
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

			tile, ok := grid.FromNative(tc.expected.Z, pt)
			if !ok {
				t.Fatal("expected ok")
			}

			if *tc.expected != *tile {
				t.Errorf("got %v expected %v", *tile, *tc.expected)
			}
		}
	}
	
	// expected = tile column, tile row
	tests := map[string]tcase{
		"3857_z0": {
			point:    geom.Point{0.0, 0.0},
			srid:     3857,
			expected: NewTile(0, 0, 0),
		},
		"3857_z0_random": {
			point:    geom.Point{96.7283, 43.5473},
			srid:     3857,
			expected: NewTile(0, 0, 0),
		},
		"3857_z10_quad1": {
			point:    geom.Point{179.99999, 85.0511},
			srid:     3857,
			expected: NewTile(10, 1023, 0),
		},
		"3857_z10_quad2": {
			point:    geom.Point{-179.99999, 85.0511},
			srid:     3857,
			expected: NewTile(10, 0, 0),
		},
		"3857_z10_quad3": {
			point:    geom.Point{-179.99999, -85.0511},
			srid:     3857,
			expected: NewTile(10, 0, 1023),
		},
		"3857_z10_quad4": {
			point:    geom.Point{179.99999, -85.0511},
			srid:     3857,
			expected: NewTile(10, 1023, 1023),
		},
		"4326_z0_quad1": {
			point:    geom.Point{0.0, 0.0},
			srid:     4326,
			expected: NewTile(0, 1, 0),
		},
		"4326_z0_quad2": {
			point:    geom.Point{-1.0, 0.0},
			srid:     4326,
			expected: NewTile(0, 0, 0),
		},
		"4326_z10_quad1": {
			point:    geom.Point{179.99999, 89.99999},
			srid:     4326,
			expected: NewTile(10, 2047, 0),
		},
		"4326_z10_quad2": {
			point:    geom.Point{-179.99999, 89.99999},
			srid:     4326,
			expected: NewTile(10, 0, 0),
		},
		"4326_z10_quad3": {
			point:    geom.Point{-179.99999, -89.99999},
			srid:     4326,
			expected: NewTile(10, 0, 1023),
		},
		"4326_z10_quad4": {
			point:    geom.Point{179.99999, -89.99999},
			srid:     4326,
			expected: NewTile(10, 2047, 1023),
		},
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}

func TestToNative(t *testing.T) {
	type tcase struct {
		tile     *Tile
		srid     uint
		expected geom.Point
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			grid, err := NewGrid(tc.srid)
			if err != nil {
				t.Fatal(err)
			}

			pt, ok := grid.ToNative(tc.tile)
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

			if pt != tc.expected {
				t.Errorf("got %v expected %v", pt, tc.expected)
			}

		}
	}

	tests := map[string]tcase{
		"3857_z0": {
			tile:     NewTile(0, 0, 0),
			srid:     3857,
			expected: geom.Point{-179.9999999749438, 85.05112877764508},
		},
		"3857_z10_q1": {
			tile:     NewTile(10, 1023, 0),
			srid:     3857,
			expected: geom.Point{179.64843747499273, 85.05112877764508},
		},
		"3857_z10_q2": {
			tile:     NewTile(10, 0, 0),
			srid:     3857,
			expected: geom.Point{-179.9999999749438, 85.05112877764508},
		},
		"3857_z10_q3": {
			tile:     NewTile(10, 0, 1023),
			srid:     3857,
			expected: geom.Point{-179.9999999749438, -85.0207077409554},
		},
		"3857_z10_q4": {
			tile:     NewTile(10, 1023, 1023),
			srid:     3857,
			expected: geom.Point{179.64843747499273, -85.0207077409554},
		},
		"4326_z0_q1": {
			tile:     NewTile(0, 1, 0),
			srid:     4326,
			expected: geom.Point{0, 90},
		},
		"4326_z0_q2": {
			tile:     NewTile(0, 0, 0),
			srid:     4326,
			expected: geom.Point{-180, 90},
		},
		"4326_z10_q1": {
			tile:     NewTile(10, 2047, 0),
			srid:     4326,
			expected: geom.Point{179.82421875, 90},
		},
		"4326_z10_q2": {
			tile:     NewTile(10, 0, 0),
			srid:     4326,
			expected: geom.Point{-180, 90},
		},
		"4326_z10_q3": {
			tile:     NewTile(10, 0, 1023),
			srid:     4326,
			expected: geom.Point{-180, -89.82421875},
		},
		"4326_z10_q4": {
			tile:     NewTile(10, 2047, 1023),
			srid:     4326,
			expected: geom.Point{179.82421875, -89.82421875},
		},
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}
