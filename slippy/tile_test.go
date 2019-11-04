package slippy_test

import (
	"math"
	"strconv"
	"testing"

	"github.com/go-spatial/proj"

	"github.com/go-spatial/geom/spherical"

	"reflect"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/slippy"
)

func TestNewTile(t *testing.T) {
	type tcase struct {
		z, x, y  uint
		buffer   float64
		srid     uint
		eBounds  *geom.Extent
		eExtent  *geom.Extent
		eBExtent *geom.Extent
	}
	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {

			// Test the new functions.
			tile := slippy.NewTile(tc.z, tc.x, tc.y)
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
			}
			{
				bounds := tile.Extent4326(tc.srid)
				bitTolerence2 := int64(math.Float64bits(1.01) - math.Float64bits(1.00))
				for i := 0; i < 4; i++ {
					if !cmp.Float64(bounds[i], tc.eBounds[i], 0.01, bitTolerence2) {
						t.Errorf("bounds[%v] , expected %v got %v", i, tc.eBounds[i], bounds[i])

					}
				}
			}
			{
				bufferedExtent := tile.NativeExtent(tc.srid).ExpandBy(slippy.PixelsToProjectedUnits(tile.Z, uint(tc.buffer), tc.srid))

				if !cmp.GeomExtent(tc.eBExtent, bufferedExtent) {
					t.Errorf("buffered extent, expected %v got %v", tc.eBExtent, bufferedExtent)
				}
			}
			{
				extent := tile.NativeExtent(tc.srid)

				if !cmp.GeomExtent(tc.eExtent, extent) {
					t.Errorf("extent, expected %v got %v", tc.eExtent, extent)
				}
			}

		}
	}
	tests := [...]tcase{
		{
			z:      2,
			x:      1,
			y:      1,
			buffer: 64,
			srid:   proj.WebMercator,
			eExtent: &geom.Extent{
				-10018754.17, 0,
				0, 10018754.17,
			},
			eBExtent: &geom.Extent{
				-1.017529720390625e+07, -156543.03390624933,
				156543.03390624933, 1.017529720390625e+07,
			},
			eBounds: spherical.Hull(
				[2]float64{-90, 66.51},
				[2]float64{0, 0},
			),
		},
		{
			z:      16,
			x:      11436,
			y:      26461,
			buffer: 64,
			srid:   proj.WebMercator,
			eExtent: &geom.Extent{
				-13044437.497219238996, 3856095.202393799,
				-13043826.000993041, 3856706.6986199953,
			},
			eBExtent: &geom.Extent{
				-1.3044447051847773e+07, 3.856085647765265e+06,
				-1.3043816446364507e+07, 3.8567162532485295e+06,
			},
			eBounds: spherical.Hull(
				[2]float64{-117.18, 32.70},
				[2]float64{-117.17, 32.70},
			),
		},
		{
			z:      0,
			x:      0,
			y:      0,
			buffer: 64,
			srid:   4326,
			eExtent: &geom.Extent{
				-180, -90,
				0, 90,
			},
			eBExtent: &geom.Extent{
				-182.8125, -92.8125,
				2.8125, 92.8125,
			},
			eBounds: spherical.Hull(
				[2]float64{-180, -90},
				[2]float64{0, 90},
			),
		},
	}
	for i, tc := range tests {
		t.Run(strconv.FormatUint(uint64(i), 10), fn(tc))
	}

}

func TestNewTileLatLon(t *testing.T) {
	type tcase struct {
		z, x, y  uint
		lat, lon float64
		buffer   float64
		srid     uint
	}
	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {

			// Test the new functions.
			tile := slippy.NewTileLatLon(tc.z, tc.lat, tc.lon, tc.srid)
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
			srid:   3857,
		},
		"center": {
			z:      8,
			x:      128,
			y:      128,
			lat:    0,
			lon:    0,
			buffer: 64,
			srid:   3857,
		},
		"arbitrary zoom 2": {
			z:      2,
			x:      2,
			y:      3,
			lat:    -70,
			lon:    20,
			buffer: 64,
			srid:   3857,
		},
		"arbitrary zoom 16": {
			z:      16,
			x:      11436,
			y:      26461,
			lat:    32.705,
			lon:    -117.176,
			buffer: 64,
			srid:   3857,
		},
	}

	for k, tc := range tests {
		t.Run(k, fn(tc))
	}
}

func TestRangeFamilyAt(t *testing.T) {
	type coord struct {
		z, x, y uint
	}

	type tcase struct {
		tile     *slippy.Tile
		tileSRID uint
		zoomAt   uint
		expected []coord
	}

	isIn := func(arr []coord, c coord) bool {
		for _, v := range arr {
			if v == c {
				return true
			}
		}

		return false
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {

			coordList := make([]coord, 0, len(tc.expected))
			tc.tile.RangeFamilyAt(tc.zoomAt, tc.tileSRID, func(tile *slippy.Tile, srid uint) error {
				z, x, y := tile.ZXY()
				c := coord{z, x, y}

				coordList = append(coordList, c)

				return nil
			})

			if len(coordList) != len(tc.expected) {
				t.Fatalf("coordinate list length, expected %d, got %d", len(tc.expected), len(coordList))
			}

			for _, v := range tc.expected {
				if !isIn(coordList, v) {
					t.Logf("coordinates: %v", coordList)
					t.Fatalf("coordinate exists, expected %v,  got missing", v)
				}
			}

		}
	}

	testcases := map[string]tcase{
		"children 1": {
			tile:     slippy.NewTile(0, 0, 0),
			tileSRID: proj.WebMercator,
			zoomAt:   1,
			expected: []coord{
				{1, 0, 0},
				{1, 0, 1},
				{1, 1, 0},
				{1, 1, 1},
			},
		},
		"children 2": {
			tile:     slippy.NewTile(8, 3, 5),
			tileSRID: proj.WebMercator,
			zoomAt:   10,
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
			tile:     slippy.NewTile(1, 0, 0),
			tileSRID: proj.WebMercator,
			zoomAt:   0,
			expected: []coord{
				{0, 0, 0},
			},
		},
		"parent 2": {
			tile:     slippy.NewTile(3, 3, 5),
			tileSRID: proj.WebMercator,
			zoomAt:   1,
			expected: []coord{
				{1, 0, 1},
			},
		},
		"parent 4326 1": {
			tile:     slippy.NewTile(1, 3, 0),
			tileSRID: 4326,
			zoomAt:   0,
			expected: []coord{
				{0, 1, 0},
			},
		},
		"parent 4326 2": {
			tile:     slippy.NewTile(4, 31, 15),
			tileSRID: 4326,
			zoomAt:   1,
			expected: []coord{
				{1, 3, 1},
			},
		},
		"children 4326": {
			tile:     slippy.NewTile(2, 7, 3),
			tileSRID: 4326,
			zoomAt:   3,
			expected: []coord{
				{3, 14, 6},
				{3, 15, 6},
				{3, 14, 7},
				{3, 15, 7},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, fn(tc))
	}
}

func TestNewTileMinMaxer(t *testing.T) {
	type tcase struct {
		mm       geom.MinMaxer
		tile     *slippy.Tile
		tileSRID uint
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {

			tile := slippy.NewTileMinMaxer(tc.mm, tc.tileSRID)
			if !reflect.DeepEqual(tile, tc.tile) {
				t.Errorf("tile, expected %v, got %v", tc.tile, tile)
			}

		}
	}

	testcases := map[string]tcase{
		"1": {
			mm: spherical.Hull(
				[2]float64{-179.0, 85.0},
				[2]float64{179.0, -85.0}),
			tile:     slippy.NewTile(0, 0, 0),
			tileSRID: proj.WebMercator,
		},
		"2": {
			mm:       slippy.NewTile(15, 2, 98).Extent4326(3857),
			tile:     slippy.NewTile(15, 2, 98),
			tileSRID: proj.WebMercator,
		},
	}

	for name, tc := range testcases {
		t.Run(name, fn(tc))
	}
}

func TestFromBounds(t *testing.T) {

	type tcase struct {
		Bounds   *geom.Extent
		Z        uint
		Tiles    []slippy.Tile
		TileSRID uint
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {

			tiles := slippy.FromBounds(tc.Bounds, tc.Z, 3857)
			if !reflect.DeepEqual(tiles, tc.Tiles) {
				t.Errorf("tiles, expected %v, got %v", tc.Tiles, tiles)
			}

		}
	}
	tests := map[string]tcase{
		"nil bounds": tcase{},
		"San Diego 15z": tcase{
			Z:        15,
			Bounds:   spherical.Hull([2]float64{-117.15, 32.6894743}, [2]float64{-116.804, 32.6339}),
			TileSRID: 3857,
			Tiles: []slippy.Tile{
				{Z: 15, X: 5720, Y: 13232}, {Z: 15, X: 5720, Y: 13233}, {Z: 15, X: 5720, Y: 13234}, {Z: 15, X: 5720, Y: 13235}, {Z: 15, X: 5720, Y: 13236},
				{Z: 15, X: 5720, Y: 13237}, {Z: 15, X: 5720, Y: 13238}, {Z: 15, X: 5721, Y: 13232}, {Z: 15, X: 5721, Y: 13233}, {Z: 15, X: 5721, Y: 13234},
				{Z: 15, X: 5721, Y: 13235}, {Z: 15, X: 5721, Y: 13236}, {Z: 15, X: 5721, Y: 13237}, {Z: 15, X: 5721, Y: 13238}, {Z: 15, X: 5722, Y: 13232},
				{Z: 15, X: 5722, Y: 13233}, {Z: 15, X: 5722, Y: 13234}, {Z: 15, X: 5722, Y: 13235}, {Z: 15, X: 5722, Y: 13236}, {Z: 15, X: 5722, Y: 13237},
				{Z: 15, X: 5722, Y: 13238}, {Z: 15, X: 5723, Y: 13232}, {Z: 15, X: 5723, Y: 13233}, {Z: 15, X: 5723, Y: 13234}, {Z: 15, X: 5723, Y: 13235},
				{Z: 15, X: 5723, Y: 13236}, {Z: 15, X: 5723, Y: 13237}, {Z: 15, X: 5723, Y: 13238}, {Z: 15, X: 5724, Y: 13232}, {Z: 15, X: 5724, Y: 13233},
				{Z: 15, X: 5724, Y: 13234}, {Z: 15, X: 5724, Y: 13235}, {Z: 15, X: 5724, Y: 13236}, {Z: 15, X: 5724, Y: 13237}, {Z: 15, X: 5724, Y: 13238},
				{Z: 15, X: 5725, Y: 13232}, {Z: 15, X: 5725, Y: 13233}, {Z: 15, X: 5725, Y: 13234}, {Z: 15, X: 5725, Y: 13235}, {Z: 15, X: 5725, Y: 13236},
				{Z: 15, X: 5725, Y: 13237}, {Z: 15, X: 5725, Y: 13238}, {Z: 15, X: 5726, Y: 13232}, {Z: 15, X: 5726, Y: 13233}, {Z: 15, X: 5726, Y: 13234},
				{Z: 15, X: 5726, Y: 13235}, {Z: 15, X: 5726, Y: 13236}, {Z: 15, X: 5726, Y: 13237}, {Z: 15, X: 5726, Y: 13238}, {Z: 15, X: 5727, Y: 13232},
				{Z: 15, X: 5727, Y: 13233}, {Z: 15, X: 5727, Y: 13234}, {Z: 15, X: 5727, Y: 13235}, {Z: 15, X: 5727, Y: 13236}, {Z: 15, X: 5727, Y: 13237},
				{Z: 15, X: 5727, Y: 13238}, {Z: 15, X: 5728, Y: 13232}, {Z: 15, X: 5728, Y: 13233}, {Z: 15, X: 5728, Y: 13234}, {Z: 15, X: 5728, Y: 13235},
				{Z: 15, X: 5728, Y: 13236}, {Z: 15, X: 5728, Y: 13237}, {Z: 15, X: 5728, Y: 13238}, {Z: 15, X: 5729, Y: 13232}, {Z: 15, X: 5729, Y: 13233},
				{Z: 15, X: 5729, Y: 13234}, {Z: 15, X: 5729, Y: 13235}, {Z: 15, X: 5729, Y: 13236}, {Z: 15, X: 5729, Y: 13237}, {Z: 15, X: 5729, Y: 13238},
				{Z: 15, X: 5730, Y: 13232}, {Z: 15, X: 5730, Y: 13233}, {Z: 15, X: 5730, Y: 13234}, {Z: 15, X: 5730, Y: 13235}, {Z: 15, X: 5730, Y: 13236},
				{Z: 15, X: 5730, Y: 13237}, {Z: 15, X: 5730, Y: 13238}, {Z: 15, X: 5731, Y: 13232}, {Z: 15, X: 5731, Y: 13233}, {Z: 15, X: 5731, Y: 13234},
				{Z: 15, X: 5731, Y: 13235}, {Z: 15, X: 5731, Y: 13236}, {Z: 15, X: 5731, Y: 13237}, {Z: 15, X: 5731, Y: 13238}, {Z: 15, X: 5732, Y: 13232},
				{Z: 15, X: 5732, Y: 13233}, {Z: 15, X: 5732, Y: 13234}, {Z: 15, X: 5732, Y: 13235}, {Z: 15, X: 5732, Y: 13236}, {Z: 15, X: 5732, Y: 13237},
				{Z: 15, X: 5732, Y: 13238}, {Z: 15, X: 5733, Y: 13232}, {Z: 15, X: 5733, Y: 13233}, {Z: 15, X: 5733, Y: 13234}, {Z: 15, X: 5733, Y: 13235},
				{Z: 15, X: 5733, Y: 13236}, {Z: 15, X: 5733, Y: 13237}, {Z: 15, X: 5733, Y: 13238}, {Z: 15, X: 5734, Y: 13232}, {Z: 15, X: 5734, Y: 13233},
				{Z: 15, X: 5734, Y: 13234}, {Z: 15, X: 5734, Y: 13235}, {Z: 15, X: 5734, Y: 13236}, {Z: 15, X: 5734, Y: 13237}, {Z: 15, X: 5734, Y: 13238},
				{Z: 15, X: 5735, Y: 13232}, {Z: 15, X: 5735, Y: 13233}, {Z: 15, X: 5735, Y: 13234}, {Z: 15, X: 5735, Y: 13235}, {Z: 15, X: 5735, Y: 13236},
				{Z: 15, X: 5735, Y: 13237}, {Z: 15, X: 5735, Y: 13238}, {Z: 15, X: 5736, Y: 13232}, {Z: 15, X: 5736, Y: 13233}, {Z: 15, X: 5736, Y: 13234},
				{Z: 15, X: 5736, Y: 13235}, {Z: 15, X: 5736, Y: 13236}, {Z: 15, X: 5736, Y: 13237}, {Z: 15, X: 5736, Y: 13238}, {Z: 15, X: 5737, Y: 13232},
				{Z: 15, X: 5737, Y: 13233}, {Z: 15, X: 5737, Y: 13234}, {Z: 15, X: 5737, Y: 13235}, {Z: 15, X: 5737, Y: 13236}, {Z: 15, X: 5737, Y: 13237},
				{Z: 15, X: 5737, Y: 13238}, {Z: 15, X: 5738, Y: 13232}, {Z: 15, X: 5738, Y: 13233}, {Z: 15, X: 5738, Y: 13234}, {Z: 15, X: 5738, Y: 13235},
				{Z: 15, X: 5738, Y: 13236}, {Z: 15, X: 5738, Y: 13237}, {Z: 15, X: 5738, Y: 13238}, {Z: 15, X: 5739, Y: 13232}, {Z: 15, X: 5739, Y: 13233},
				{Z: 15, X: 5739, Y: 13234}, {Z: 15, X: 5739, Y: 13235}, {Z: 15, X: 5739, Y: 13236}, {Z: 15, X: 5739, Y: 13237}, {Z: 15, X: 5739, Y: 13238},
				{Z: 15, X: 5740, Y: 13232}, {Z: 15, X: 5740, Y: 13233}, {Z: 15, X: 5740, Y: 13234}, {Z: 15, X: 5740, Y: 13235}, {Z: 15, X: 5740, Y: 13236},
				{Z: 15, X: 5740, Y: 13237}, {Z: 15, X: 5740, Y: 13238}, {Z: 15, X: 5741, Y: 13232}, {Z: 15, X: 5741, Y: 13233}, {Z: 15, X: 5741, Y: 13234},
				{Z: 15, X: 5741, Y: 13235}, {Z: 15, X: 5741, Y: 13236}, {Z: 15, X: 5741, Y: 13237}, {Z: 15, X: 5741, Y: 13238}, {Z: 15, X: 5742, Y: 13232},
				{Z: 15, X: 5742, Y: 13233}, {Z: 15, X: 5742, Y: 13234}, {Z: 15, X: 5742, Y: 13235}, {Z: 15, X: 5742, Y: 13236}, {Z: 15, X: 5742, Y: 13237},
				{Z: 15, X: 5742, Y: 13238}, {Z: 15, X: 5743, Y: 13232}, {Z: 15, X: 5743, Y: 13233}, {Z: 15, X: 5743, Y: 13234}, {Z: 15, X: 5743, Y: 13235},
				{Z: 15, X: 5743, Y: 13236}, {Z: 15, X: 5743, Y: 13237}, {Z: 15, X: 5743, Y: 13238}, {Z: 15, X: 5744, Y: 13232}, {Z: 15, X: 5744, Y: 13233},
				{Z: 15, X: 5744, Y: 13234}, {Z: 15, X: 5744, Y: 13235}, {Z: 15, X: 5744, Y: 13236}, {Z: 15, X: 5744, Y: 13237}, {Z: 15, X: 5744, Y: 13238},
				{Z: 15, X: 5745, Y: 13232}, {Z: 15, X: 5745, Y: 13233}, {Z: 15, X: 5745, Y: 13234}, {Z: 15, X: 5745, Y: 13235}, {Z: 15, X: 5745, Y: 13236},
				{Z: 15, X: 5745, Y: 13237}, {Z: 15, X: 5745, Y: 13238}, {Z: 15, X: 5746, Y: 13232}, {Z: 15, X: 5746, Y: 13233}, {Z: 15, X: 5746, Y: 13234},
				{Z: 15, X: 5746, Y: 13235}, {Z: 15, X: 5746, Y: 13236}, {Z: 15, X: 5746, Y: 13237}, {Z: 15, X: 5746, Y: 13238}, {Z: 15, X: 5747, Y: 13232},
				{Z: 15, X: 5747, Y: 13233}, {Z: 15, X: 5747, Y: 13234}, {Z: 15, X: 5747, Y: 13235}, {Z: 15, X: 5747, Y: 13236}, {Z: 15, X: 5747, Y: 13237},
				{Z: 15, X: 5747, Y: 13238}, {Z: 15, X: 5748, Y: 13232}, {Z: 15, X: 5748, Y: 13233}, {Z: 15, X: 5748, Y: 13234}, {Z: 15, X: 5748, Y: 13235},
				{Z: 15, X: 5748, Y: 13236}, {Z: 15, X: 5748, Y: 13237}, {Z: 15, X: 5748, Y: 13238}, {Z: 15, X: 5749, Y: 13232}, {Z: 15, X: 5749, Y: 13233},
				{Z: 15, X: 5749, Y: 13234}, {Z: 15, X: 5749, Y: 13235}, {Z: 15, X: 5749, Y: 13236}, {Z: 15, X: 5749, Y: 13237}, {Z: 15, X: 5749, Y: 13238},
				{Z: 15, X: 5750, Y: 13232}, {Z: 15, X: 5750, Y: 13233}, {Z: 15, X: 5750, Y: 13234}, {Z: 15, X: 5750, Y: 13235}, {Z: 15, X: 5750, Y: 13236},
				{Z: 15, X: 5750, Y: 13237}, {Z: 15, X: 5750, Y: 13238}, {Z: 15, X: 5751, Y: 13232}, {Z: 15, X: 5751, Y: 13233}, {Z: 15, X: 5751, Y: 13234},
				{Z: 15, X: 5751, Y: 13235}, {Z: 15, X: 5751, Y: 13236}, {Z: 15, X: 5751, Y: 13237}, {Z: 15, X: 5751, Y: 13238}, {Z: 15, X: 5752, Y: 13232},
				{Z: 15, X: 5752, Y: 13233}, {Z: 15, X: 5752, Y: 13234}, {Z: 15, X: 5752, Y: 13235}, {Z: 15, X: 5752, Y: 13236}, {Z: 15, X: 5752, Y: 13237},
				{Z: 15, X: 5752, Y: 13238},
			},
		},
		"San Diego 11z": tcase{
			Z:        11,
			Bounds:   spherical.Hull([2]float64{-117.15, 32.6894743}, [2]float64{-116.804, 32.6339}),
			TileSRID: 3857,
			Tiles:    []slippy.Tile{{Z: 11, X: 357, Y: 827}, {Z: 11, X: 358, Y: 827}, {Z: 11, X: 359, Y: 827}},
		},
		"San Diego 9z": tcase{
			Z:        9,
			Bounds:   spherical.Hull([2]float64{-117.15, 32.6894743}, [2]float64{-116.804, 32.6339}),
			TileSRID: 3857,
			Tiles:    []slippy.Tile{{Z: 9, X: 89, Y: 206}},
		},
	}
	for name, tc := range tests {
		t.Run(name, fn(tc))
	}

}
