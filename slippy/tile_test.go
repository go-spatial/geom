package slippy

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"unicode"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/spherical"
	"github.com/go-spatial/proj"
)

var (
	//go:embed testdata
	testdata embed.FS
)

func must[T any](a T, err error) T {
	if err != nil {
		panic(err)
	}
	return a
}

func tcCurry[T any](fn func(T, *testing.T)) func(T) func(*testing.T) {
	return func(tc T) func(*testing.T) {
		return func(t *testing.T) {
			fn(tc, t)
		}
	}
}

type FindableTile struct {
	Tile
	found bool
}

func tilesFromSlice(coords []int) ([]FindableTile, error) {
	if len(coords)%3 != 0 {
		return nil, fmt.Errorf("tilesFromSlice expects an number of coordinates to be a multiple of 3: got %v", len(coords))
	}
	tiles := make([]FindableTile, 0, len(coords)/3)
	for i := 0; i < len(coords); i += 3 {
		tiles = append(tiles, FindableTile{Tile: Tile{Z: Zoom(coords[i]), X: uint(coords[i+1]), Y: uint(coords[i+2])}})
	}
	return tiles, nil
}

func LoadCoords(bytes []byte) ([]int, error) {
	// file should be a set of 3 integers separated by space, representing a tile.
	// comments are `#` followed by the comment till the end of the line.
	var (
		text         = string(bytes)
		inComment    = false
		numberBuffer = make([]rune, 0, 10)
		line         = 1
		offset       = 0
		aNum         int64
		err          error
		coords       []int
	)
	for _, aChar := range text {
		offset++
		if inComment {
			// check to see if we have a newline
			if aChar == '\n' {
				line++
				offset = 0
				inComment = false
			}
			continue
		}
		if unicode.IsSpace(aChar) || aChar == ',' || aChar == '/' || aChar == '\n' || aChar == '#' {
			if aChar == '\n' {
				line++
				offset = 0
			}

			if aChar == '#' {
				// we will be in a comment after this
				inComment = true
			}

			// found a separator character
			if len(numberBuffer) == 0 {
				continue
			}
			aNum, err = strconv.ParseInt(string(numberBuffer), 0, 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing number on line %v offset %v: %w", line, offset, err)
			}
			// reset our buffer
			numberBuffer = numberBuffer[:0]
			coords = append(coords, int(aNum))
			continue
		}
		if unicode.IsDigit(aChar) {
			numberBuffer = append(numberBuffer, aChar)
		}
	}
	if len(numberBuffer) != 0 {
		aNum, err = strconv.ParseInt(string(numberBuffer), 0, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing number on line %v offset %v: %w", line, offset, err)
		}
		numberBuffer = numberBuffer[:0]
		coords = append(coords, int(aNum))
	}
	return coords, nil
}
func LoadTiles(bytes []byte) ([]FindableTile, error) {
	coords, err := LoadCoords(bytes)
	if err != nil {
		return nil, err
	}
	return tilesFromSlice(coords)
}

func WriteTilesToFile[S fmt.Stringer](filename string, tiles []S) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	for _, tile := range tiles {
		if _, err = file.WriteString(tile.String() + "\n"); err != nil {
			return err
		}
	}
	return nil
}

func TestRangeFamilyAt2(t *testing.T) {

	type tcase struct {
		Tile
		Zoom     Zoom
		expected []int
	}

	fn := tcCurry(func(tc tcase, t *testing.T) {

		count := 0
		expectedTiles, err := tilesFromSlice(tc.expected)
		if err != nil {
			t.Fatalf("tilesFromSlice error: %v", err)
		}
		RangeFamilyAt(tc.Tile, tc.Zoom, func(tile Tile) bool {
			count++
			t.Logf("Got tile %v", tile)

			found := false
			for i := range expectedTiles {
				if !expectedTiles[i].Equal(tile) {
					continue
				}
				found = true
				expectedTiles[i].found = true
				break
			}
			if !found {
				t.Errorf("tile %v not found in expected tiles", tile)
			}
			return true
		})

		for _, v := range expectedTiles {
			if !v.found {
				t.Errorf("expected tile %v not found", v)
			}
		}

		if count != len(expectedTiles) {
			t.Fatalf("list length, expected %d, got %d", len(expectedTiles), count)
		}
	})

	testcases := map[string]tcase{
		"children 1": {
			Tile: Tile{0, 0, 0},
			Zoom: 1,
			expected: []int{
				1, 0, 0,
				1, 0, 1,
				1, 1, 0,
				1, 1, 1,
			},
		},
		"children 2": {
			Tile: Tile{8, 3, 5},
			Zoom: 10,
			expected: []int{
				10, 12, 20,
				10, 12, 21,
				10, 12, 22,
				10, 12, 23,
				//
				10, 13, 20,
				10, 13, 21,
				10, 13, 22,
				10, 13, 23,
				//
				10, 14, 20,
				10, 14, 21,
				10, 14, 22,
				10, 14, 23,
				//
				10, 15, 20,
				10, 15, 21,
				10, 15, 22,
				10, 15, 23,
			},
		},
		"parent 1": {
			Tile:     Tile{1, 0, 0},
			Zoom:     0,
			expected: []int{0, 0, 0},
		},
		"parent 2": {
			Tile:     Tile{3, 3, 5},
			Zoom:     1,
			expected: []int{1, 0, 1},
		},
		"parent 1.1 ": {
			Tile:     Tile{1, 3, 0},
			Zoom:     0,
			expected: []int{0, 1, 0},
		},
		"parent 2.1": {
			Tile:     Tile{4, 31, 15},
			Zoom:     1,
			expected: []int{1, 3, 1},
		},
		"children 1.1": {
			Tile: Tile{2, 7, 3},
			Zoom: 3,
			expected: []int{
				3, 14, 6,
				3, 15, 6,
				3, 14, 7,
				3, 15, 7,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, fn(tc))
	}
}

func TestFromBounds2(t *testing.T) {
	type tcase struct {
		Grid   TileGridder // if nil, we will default to Grid4326
		Bounds geom.PtMinMaxer
		Z      Zoom
		Tiles  []FindableTile

		err error
	}

	fn := tcCurry(func(tc tcase, t *testing.T) {
		if tc.Grid == nil {
			tc.Grid = Grid4326{}
		}

		tiles, err := FromBounds(tc.Grid, tc.Bounds, tc.Z)
		if tc.err != nil {
			if err == nil {
				t.Errorf("error expected %v, got nil", tc.err)
				return
			}
			if !errors.Is(err, tc.err) {
				t.Errorf("error expected %v, got %v", tc.err, err)
				return
			}
			return
		}
		if err != nil {
			t.Errorf("error expected nil, got %v", err)
			return
		}
		defer func() {
			if !t.Failed() {
				return
			}
			// we failed the test, let's dump the tiles
			var (
				filename = fmt.Sprintf("testdata/failed_output/%v.failed.tiles", filepath.Base(t.Name()))
				err      error
			)
			if err = WriteTilesToFile(filename, tiles); err != nil {
				t.Logf("failed to write failed tiles file(%v): %v", filename, err)
				return
			} else {
				t.Logf("wrote failed tiles file(%v)", filename)
			}
			filename = fmt.Sprintf("testdata/failed_output/%v.expected.tiles", filepath.Base(t.Name()))
			if err = WriteTilesToFile(filename, tc.Tiles); err != nil {
				t.Logf("failed to write expected tiles file(%v): %v", filename, err)
				return
			} else {
				t.Logf("wrote expected tiles file(%v)", filename)
			}
		}()
		if len(tiles) != len(tc.Tiles) {
			t.Errorf("len expected %v, got %v", len(tc.Tiles), len(tiles))
			return
		}
	NextTile:
		for i := range tiles {
			for j, tile := range tc.Tiles {
				if tile.found {
					continue
				}
				if tile.Equal(tiles[i]) {
					tc.Tiles[j].found = true
					continue NextTile
				}
			}
			t.Errorf("tile[%v] %v not found in expected tiles", i, tiles[i])
		}
		for i, tile := range tc.Tiles {
			if tile.found {
				continue
			}
			t.Errorf("expected tile[%v] %v not found in tiles", i, tc.Tiles[i])
		}

	})

	tests := map[string]tcase{
		"nil bounds": {err: ErrNilBounds},
		"San Diego 15z": {
			Z:      15,
			Bounds: spherical.Hull([2]float64{-117.15, 32.6894743}, [2]float64{-116.804, 32.6339}),
			Tiles:  must(LoadTiles(must(testdata.ReadFile("testdata/for_bounds/san_diego_15z.coords")))),
		},
		"San Diego 11z": {
			Z:      11,
			Bounds: spherical.Hull([2]float64{-117.15, 32.6894743}, [2]float64{-116.804, 32.6339}),
			Tiles:  must(LoadTiles(must(testdata.ReadFile("testdata/for_bounds/san_diego_11z.coords")))),
		},
		"San Diego 9z": {
			Z:      9,
			Bounds: spherical.Hull([2]float64{-117.15, 32.6894743}, [2]float64{-116.804, 32.6339}),
			Tiles:  must(LoadTiles(must(testdata.ReadFile("testdata/for_bounds/san_diego_9z.coords")))),
		},
		"tegola issue 997": {
			Z: 7,
			Bounds: spherical.Hull(
				[2]float64{
					2.636719,
					50.625073,
				},
				[2]float64{
					7.613525,
					53.820112,
				}),
			Tiles: must(LoadTiles(must(testdata.ReadFile("testdata/for_bounds/tegola_issue_997.coords")))),
		},
		"tegola issue 997 w seeding bounds": {
			Z:    7,
			Grid: NewGrid(proj.EPSG4326, 0),
			Bounds: spherical.Hull(
				[2]float64{
					3.011234,
					50.16669,
				},
				[2]float64{
					7.64906,
					54.683876,
				}),
			Tiles: must(LoadTiles(must(testdata.ReadFile("testdata/for_bounds/tegola_issue_997_w_seeding_bounds.coords")))),
		},
		"tegola issue 997 SRID 3857": {
			Z:    7,
			Grid: NewGrid(proj.EPSG3857, DefaultTileSize),
			Bounds: spherical.Hull(
				[2]float64{
					293518.1886,
					6555239.5457,
				},
				[2]float64{
					847533.7696,
					7136160.9607,
				}),
			Tiles: must(LoadTiles(must(testdata.ReadFile("testdata/for_bounds/tegola_issue_997.coords")))),
		},
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}

}
