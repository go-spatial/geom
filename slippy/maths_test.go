package slippy

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func Test_lat2Num(t *testing.T) {

	type tcase struct {
		TileSize uint32
		Z        Zoom
		Lat      float64

		y int
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			// this is to help understand things.
			lat := y2deg(tc.Z, tc.y)
			t.Logf("z: %v y: %v lat = %v tc.Lat: %v", tc.Z, tc.y, lat, tc.Lat)
			y := lat2Num(tc.TileSize, tc.Z, tc.Lat)
			if y != tc.y {
				t.Errorf("y got %d expected %d", y, tc.y)
			}
		}
	}
	tests := []tcase{
		{
			Lat: 38.889814,
			Z:   11,
			y:   783,
		},
		{
			Lat: 38.889814,
			y:   0,
		},
		{
			Lat: -86,
			y:   0,
		},
		{
			Lat: -Lat4326Max,
			Z:   0,
			y:   0,
		},
		{
			Lat: -85.0511,
			Z:   1,
			y:   1,
		},
		{
			// example from orb/maptile
			Lat: 41.850033,
			Z:   28,
			y:   99798110,
		},
		{ // example from open street maps slippy tile
			//	Lat: Radians2Degree(0.66693624687),
			Lat: 35.6590699, // 4326
			Z:   18,
			y:   103246,
		},
	}
	for i, tc := range tests {
		t.Run(fmt.Sprintf("test_%d", i), fn(tc))
	}

}

func Test_lon2Num(t *testing.T) {

	type tcase struct {
		TileSize uint32
		Z        Zoom
		Lon      float64

		x int
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			x := lon2Num(tc.TileSize, tc.Z, tc.Lon)
			if x != tc.x {
				t.Errorf("x got %d expected %d", x, tc.x)
			}
		}
	}
	tests := []tcase{
		{
			Lon: -77.035915,
			Z:   11,
			x:   585,
		},
		{
			Lon: 38.889814,
			x:   0,
		},
		{
			Lon: Lon4326Max,
			Z:   0,
			x:   0,
		},
		{
			Lon: -Lon4326Max,
			Z:   1,
			x:   0,
		},
		{
			Lon: 139.7006793,
			Z:   18,
			x:   232798,
		},
	}
	for i, tc := range tests {
		t.Run(fmt.Sprintf("test_%d", i), fn(tc))
	}
	var tc tcase
	for i := -Lon4326Max; i < Lon4326Max; i++ {
		tc.Lon = float64(i)
		tc.TileSize = 256
		t.Run(fmt.Sprintf("z0_test_%v", i), fn(tc))
	}
	for z := Zoom(0); z <= 20; z++ {
		tc.Lon = -Lon4326Max
		tc.TileSize = 256
		tc.Z = z
		t.Run(fmt.Sprintf("z%02d_test_neg_lon_max", z), fn(tc))
	}
	for z := Zoom(0); z <= 20; z++ {
		tc.Lon = Lon4326Max
		tc.TileSize = 256
		tc.Z = z
		tc.x = (1 << z) - 1 // last tile
		t.Run(fmt.Sprintf("z%02d_test_lon_max", z), fn(tc))
	}

}

// Test_RoundTrip will go ToNative with a given tile, that use the native point to go FromNative, and verify that the
// Tile is the same as the starting tile.
func Test_RoundTrip(t *testing.T) {

	type tcase struct {
		Tile    Tile
		ToErr   error
		FromErr error
	}
	fn := func(tc tcase) func(t *testing.T) {
		g := Grid4326{}
		return func(t *testing.T) {
			pt, err := g.ToNative(tc.Tile)
			if tc.ToErr != nil {
				if err == nil {
					t.Errorf("to err got nil, want %v", tc.ToErr)
					return
				}
				if !errors.Is(err, tc.ToErr) {
					t.Errorf("to err got %v, want %v", err, tc.ToErr)
					return
				}
				return
			}
			if err != nil {
				t.Errorf("to err got %v, want nil", err)
			}
			tile, err := g.FromNative(Zoom(tc.Tile.Z), pt)
			if tc.FromErr != nil {
				if err == nil {
					t.Errorf("from err got nil, want %v", tc.FromErr)
					return
				}
				if !errors.Is(err, tc.FromErr) {
					t.Errorf("from err got %v, want %v", err, tc.FromErr)
					return
				}
				return
			}
			if err != nil {
				t.Errorf("from err got %v, want nil", err)
			}
			if !reflect.DeepEqual(tile, tc.Tile) {
				t.Errorf("tile (pt: %v) got %v expected %v", pt, tile, tc.Tile)
				return
			}
		}
	}
	// Test for all tiles in Z0, we want to make sure we can over extend the tile.
	for z := Zoom(0); z <= 7; z++ {
		for x := 0; x < (1 << z); x++ {
			for y := 0; y < (1 << z); y++ {
				t.Run(fmt.Sprintf("Tile(%v,%v,%v)", z, x, y), fn(tcase{Tile: Tile{
					Z: z,
					X: uint(x),
					Y: uint(y),
				}}))
			}
		}
	}

}
