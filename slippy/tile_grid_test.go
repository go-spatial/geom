package slippy

import (
	"fmt"
	"testing"
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
			grid := GetGrid(tc.srid)
			sizex, sizey := grid.GridSize(tc.zoom)
			if sizex != tc.expectedSizeX {
				t.Errorf("testcase (%v) failed. output (%v) does not match expected (%v)", t.Name(), sizex, tc.expectedSizeX)
			}
			if sizey != tc.expectedSizeY {
				t.Errorf("testcase (%v) failed. output (%v) does not match expected (%v)", t.Name(), sizey, tc.expectedSizeY)
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
		grid := GetGrid(tc.srid)

		return func(t *testing.T) {
			output := grid.ContainsIndex(tc.zoom, tc.x, tc.y)

			if output != tc.expected {
				t.Errorf("testcase (%v) failed. output (%v) does not match expected (%v)", t.Name(), output, tc.expected)
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
			x:        1,
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
			x:        1,
			y:        0,
			expected: true,
		},
		"4326_zoom0_fail": {
			srid:     4326,
			zoom:     0,
			x:        0,
			y:        1,
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

func TestLat2YIndex(t *testing.T) {
	type tcase struct {
		lat      float64
		srid     uint
		zoom     uint
		expected uint
	}

	fn := func(tc tcase) func(t *testing.T) {
		grid := GetGrid(tc.srid)
		return func(t *testing.T) {
			output := grid.Lat2YIndex(tc.zoom, tc.lat)
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

func TestLon2XIndex(t *testing.T) {
	type tcase struct {
		lon      float64
		srid     uint
		zoom     uint
		expected uint
	}

	fn := func(tc tcase) func(t *testing.T) {
		grid := GetGrid(tc.srid)
		return func(t *testing.T) {
			output := grid.Lon2XIndex(tc.zoom, tc.lon)
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

func TestXIndex2Lon(t *testing.T) {
	type tcase struct {
		x        uint
		srid     uint
		zoom     uint
		expected float64
	}

	fn := func(tc tcase) func(t *testing.T) {
		grid := GetGrid(tc.srid)
		return func(t *testing.T) {
			output := grid.XIndex2Lon(tc.zoom, tc.x)
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

func TestYIndex2Lat(t *testing.T) {
	type tcase struct {
		y        uint
		srid     uint
		zoom     uint
		expected float64
	}

	fn := func(tc tcase) func(t *testing.T) {
		grid := GetGrid(tc.srid)
		return func(t *testing.T) {
			output := grid.YIndex2Lat(tc.zoom, tc.y)
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
