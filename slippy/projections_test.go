package slippy

import (
	"fmt"
	"testing"
)

func TestLat2Tile(t *testing.T) {
	type tcase struct {
		lat      float64
		srid     uint
		zoom     uint
		expected uint
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			output := Lat2Tile(tc.zoom, tc.lat, tc.srid)
			if output != tc.expected {
				t.Errorf("testcase (%v) failed. output (%v) does not match expected (%v)", t.Name(), output, tc.expected)
			}
		}
	}

	tests := map[string]tcase{
		"3857_0": {
			lat:      0.0,
			srid:     3857,
			zoom:     0,
			expected: 0,
		},
		"3857_south": {
			lat:      -85.0511,
			srid:     3857,
			zoom:     0,
			expected: 0,
		},
		"3857_north": {
			lat:      85.0511,
			srid:     3857,
			zoom:     0,
			expected: 0,
		},
		"3857_z10_north": {
			lat:      85.0511,
			srid:     3857,
			zoom:     10,
			expected: 0,
		},
		"3857_z10_south": {
			lat:      -85.0511,
			srid:     3857,
			zoom:     10,
			expected: 1023,
		},
		"4326_0": {
			lat:      0.0,
			srid:     4326,
			zoom:     0,
			expected: 0,
		},
		"4326_south": {
			lat:      -89.99999,
			srid:     4326,
			zoom:     0,
			expected: 0,
		},
		"4326_north": {
			lat:      89.99999,
			srid:     4326,
			zoom:     0,
			expected: 0,
		},
		"4326_z10_north": {
			lat:      89.99999,
			srid:     4326,
			zoom:     10,
			expected: 0,
		},
		"4326_z10_south": {
			lat:      -89.99999,
			srid:     4326,
			zoom:     10,
			expected: 1023,
		},
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}

func TestLon2Tile(t *testing.T) {
	type tcase struct {
		lon      float64
		srid     uint
		zoom     uint
		expected uint
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			output := Lon2Tile(tc.zoom, tc.lon, tc.srid)
			if output != tc.expected {
				t.Errorf("testcase (%v) failed. output (%v) does not match expected (%v)", t.Name(), output, tc.expected)
			}
		}
	}

	tests := map[string]tcase{
		"3857_0": {
			lon:      0.0,
			srid:     3857,
			zoom:     0,
			expected: 0,
		},
		"3857_west": {
			lon:      -179.99999,
			srid:     3857,
			zoom:     0,
			expected: 0,
		},
		"3857_east": {
			lon:      179.99999,
			srid:     3857,
			zoom:     0,
			expected: 0,
		},
		"3857_z10_west": {
			lon:      -179.99999,
			srid:     3857,
			zoom:     10,
			expected: 0,
		},
		"3857_z10_east": {
			lon:      179.99999,
			srid:     3857,
			zoom:     10,
			expected: 1023,
		},
		"4326_0": {
			lon:      0.0,
			srid:     4326,
			zoom:     0,
			expected: 1,
		},
		"4326_west": {
			lon:      -179.99999,
			srid:     4326,
			zoom:     0,
			expected: 0,
		},
		"4326_east": {
			lon:      179.99999,
			srid:     4326,
			zoom:     0,
			expected: 1,
		},
		"4326_z10_west": {
			lon:      -179.99999,
			srid:     4326,
			zoom:     10,
			expected: 0,
		},
		"4326_z10_east": {
			lon:      179.99999,
			srid:     4326,
			zoom:     10,
			expected: 2047,
		},
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}

func TestTile2Lon(t *testing.T) {
	type tcase struct {
		x        uint
		srid     uint
		zoom     uint
		expected float64
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			output := Tile2Lon(tc.zoom, tc.x, tc.srid)
			if output != tc.expected {
				t.Errorf("testcase (%v) failed. output (%v) does not match expected (%v)", t.Name(), output, tc.expected)
			}
		}
	}

	tests := map[string]tcase{
		"3857_z0_west": {
			x:        0,
			srid:     3857,
			zoom:     0,
			expected: -180,
		},
		"3857_z10_west": {
			x:        0,
			srid:     3857,
			zoom:     10,
			expected: -180,
		},
		"3857_z10_east": {
			x:        1023,
			srid:     3857,
			zoom:     10,
			expected: 179.6484375,
		},
		"4326_z0_west": {
			x:        0,
			srid:     4326,
			zoom:     0,
			expected: -180,
		},
		"4326_z0_east": {
			x:        1,
			srid:     4326,
			zoom:     0,
			expected: 0,
		},
		"4326_z10_west": {
			x:        0,
			srid:     4326,
			zoom:     10,
			expected: -180.0,
		},
		"4326_z10_east": {
			x:        2047,
			srid:     4326,
			zoom:     10,
			expected: (179.6484375 + 180.0) / 2.0,
		},
		"4326_z10_center": {
			x:        1024,
			srid:     4326,
			zoom:     10,
			expected: 0.0,
		},
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}

func TestTile2Lat(t *testing.T) {
	type tcase struct {
		y        uint
		srid     uint
		zoom     uint
		expected float64
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			output := Tile2Lat(tc.zoom, tc.y, tc.srid)
			outs := fmt.Sprintf("%.8f", output)
			if outs != fmt.Sprintf("%.8f", tc.expected) {
				t.Errorf("testcase (%v) failed. output (%v) does not match expected (%v) close enough", t.Name(), output, tc.expected)
			}
		}
	}

	tests := map[string]tcase{
		"3857_z0_north": {
			y:        0,
			srid:     3857,
			zoom:     0,
			expected: 85.05112878,
		},
		"3857_z10_north": {
			y:        0,
			srid:     3857,
			zoom:     10,
			expected: 85.05112878,
		},
		"3857_z10_south": {
			y:        1023,
			srid:     3857,
			zoom:     10,
			expected: -85.02070774,
		},
		"4326_z0_north": {
			y:        0,
			srid:     4326,
			zoom:     0,
			expected: 90,
		},
		"4326_z10_north": {
			y:        0,
			srid:     4326,
			zoom:     10,
			expected: 90,
		},
		"4326_z10_south": {
			y:        1023,
			srid:     4326,
			zoom:     10,
			expected: -89.82421875,
		},
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}
