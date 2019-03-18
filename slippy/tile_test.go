package slippy_test

import (
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/slippy"
)

func TestNewTile(t *testing.T) {
	type tcase struct {
		z, x, y  uint
		buffer   float64
		srid     uint64
		eBounds  [4]float64
		eExtent  *geom.Extent
		eBExtent *geom.Extent
	}
	fn := func(t *testing.T, tc tcase) {

		// Test the new functions.
		tile := slippy.NewTile(tc.z, tc.x, tc.y, tc.buffer, tc.srid)
		{
			gz, gx, gy := tile.ZXY()
			if gz != tc.z {
				t.Errorf("z, expected %v got %v", tc.z, gz)
			}
			if gx != tc.x {
				t.Errorf("x, expected %v got %v", tc.x, gx)
			}
			if gy != tc.y {
				t.Errorf("y, expected %v got %v", tc.y, gy)
			}
			if tile.Buffer != tc.buffer {
				t.Errorf("buffer, expected %v got %v", tc.buffer, tile.Buffer)
			}
			if tile.SRID != tc.srid {
				t.Errorf("srid, expected %v got %v", tc.srid, tile.SRID)
			}
		}
		{
			bounds := tile.Bounds()
			for i := 0; i < 4; i++ {
				if !cmp.Float64(bounds[i], tc.eBounds[i], 0.01) {
					t.Errorf("bounds[%v] , expected %v got %v", i, tc.eBounds[i], bounds[i])

				}
			}
		}
		{
			bufferedExtent, srid := tile.BufferedExtent()
			if srid != tc.srid {
				t.Errorf("buffered extent srid, expected %v got %v", tc.srid, srid)
			}

			if !cmp.GeomExtent(tc.eBExtent, bufferedExtent) {
				t.Errorf("buffered extent, expected %v got %v", tc.eBExtent, bufferedExtent)
			}
		}
		{
			extent, srid := tile.Extent()
			if srid != tc.srid {
				t.Errorf("extent srid, expected %v got %v", tc.srid, srid)
			}

			if !cmp.GeomExtent(tc.eExtent, extent) {
				t.Errorf("extent, expected %v got %v", tc.eExtent, extent)
			}
		}

	}
	tests := [...]tcase{
		{
			z:      2,
			x:      1,
			y:      1,
			buffer: 64,
			srid:   geom.WebMercator,
			eExtent: geom.NewExtent(
				[2]float64{-10018754.17, 10018754.17},
				[2]float64{0, 0},
			),
			eBExtent: geom.NewExtent(
				[2]float64{-1.017529720390625e+07, 1.017529720390625e+07},
				[2]float64{156543.03390624933, -156543.03390624933},
			),
			eBounds: [4]float64{-90, 66.51, 0, 0},
		},
		{
			z:      16,
			x:      11436,
			y:      26461,
			buffer: 64,
			srid:   geom.WebMercator,
			eExtent: geom.NewExtent(
				[2]float64{-13044437.497219238996, 3856706.6986199953},
				[2]float64{-13043826.000993041, 3856095.202393799},
			),
			eBExtent: geom.NewExtent(
				[2]float64{-1.3044447051847773e+07, 3.8567162532485295e+06},
				[2]float64{-1.3043816446364507e+07, 3.856085647765265e+06},
			),
			eBounds: [4]float64{-117.18, 32.71, -117.17, 32.70},
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(strconv.FormatUint(uint64(i), 10), func(t *testing.T) { fn(t, tc) })
	}

}

func TestNewTileLatLon(t *testing.T) {
	type tcase struct {
		z, x, y  uint
		lat, lon float64
		buffer   float64
		srid     uint64
	}
	fn := func(t *testing.T, tc tcase) {

		// Test the new functions.
		tile := slippy.NewTileLatLon(tc.z, tc.lat, tc.lon, tc.buffer, tc.srid)
		{
			gz, gx, gy := tile.ZXY()
			if gz != tc.z {
				t.Errorf("z, expected %v got %v", tc.z, gz)
			}
			if gx != tc.x {
				t.Errorf("x, expected %v got %v", tc.x, gx)
			}
			if gy != tc.y {
				t.Errorf("y, expected %v got %v", tc.y, gy)
			}
			if tile.Buffer != tc.buffer {
				t.Errorf("buffer, expected %v got %v", tc.buffer, tile.Buffer)
			}
			if tile.SRID != tc.srid {
				t.Errorf("srid, expected %v got %v", tc.srid, tile.SRID)
			}
		}

	}
	tests := map[string]tcase{
		"zero": {
			z:      0,
			x:      0,
			y:      0,
			lat:    0,
			lon:    0,
			buffer: 64,
			srid:   geom.WebMercator,
		},
		"center": {
			z:      8,
			x:      128,
			y:      128,
			lat:    0,
			lon:    0,
			buffer: 64,
			srid:   geom.WebMercator,
		},
		"arbitrary zoom 2": {
			z:      2,
			x:      2,
			y:      3,
			lat:    -70,
			lon:    20,
			buffer: 64,
			srid:   geom.WebMercator,
		},
		"arbitrary zoom 16": {
			z:      16,
			x:      11436,
			y:      26461,
			lat:    32.705,
			lon:    -117.176,
			buffer: 64,
			srid:   geom.WebMercator,
		},
	}

	for k, tc := range tests {
		tc := tc
		t.Run(k, func(t *testing.T) { fn(t, tc) })
	}
}

func TestRangeFamilyAt(t *testing.T) {
	type coord struct {
		z, x, y uint
	}

	testcases := map[string]struct {
		tile     *slippy.Tile
		zoomAt   uint
		expected []coord
	}{
		"children 1": {
			tile:   slippy.NewTile(0, 0, 0, 0, geom.WebMercator),
			zoomAt: 1,
			expected: []coord{
				{1, 0, 0},
				{1, 0, 1},
				{1, 1, 0},
				{1, 1, 1},
			},
		},
		"children 2": {
			tile:   slippy.NewTile(8, 3, 5, 0, geom.WebMercator),
			zoomAt: 10,
			expected: []coord{
				{10, 12, 20},
				{10, 12, 21},
				{10, 12, 22},
				{10, 12, 23},
				//
				{10, 13, 20},
				{10, 13, 21},
				{10, 13, 22},
				{10, 13, 23},
				//
				{10, 14, 20},
				{10, 14, 21},
				{10, 14, 22},
				{10, 14, 23},
				//
				{10, 15, 20},
				{10, 15, 21},
				{10, 15, 22},
				{10, 15, 23},
			},
		},
		"parent 1": {
			tile:   slippy.NewTile(1, 0, 0, 0, geom.WebMercator),
			zoomAt: 0,
			expected: []coord{
				{0, 0, 0},
			},
		},
		"parent 2": {
			tile:   slippy.NewTile(3, 3, 5, 0, geom.WebMercator),
			zoomAt: 1,
			expected: []coord{
				{1, 0, 1},
			},
		},
	}

	isIn := func(arr []coord, c coord) bool {
		for _, v := range arr {
			if v == c {
				return true
			}
		}

		return false
	}

	for k, tc := range testcases {
		coordList := make([]coord, 0, len(tc.expected))
		tc.tile.RangeFamilyAt(tc.zoomAt, func(tile *slippy.Tile) error {
			z, x, y := tile.ZXY()
			c := coord{z, x, y}

			coordList = append(coordList, c)

			return nil
		})

		if len(coordList) != len(tc.expected) {
			t.Fatalf("[%v] expected coordinate list of length %d, got %d", k, len(tc.expected), len(coordList))
		}

		for _, v := range tc.expected {
			if !isIn(coordList, v) {
				t.Fatalf("[%v] expected coordinate %v missing from list %v", k, v, coordList)
			}
		}
	}
}
