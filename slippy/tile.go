package slippy

import (
	"errors"
	"fmt"
	"math"

	"github.com/go-spatial/proj"

	"github.com/go-spatial/geom"
)

// MaxZoom is the lowest zoom (furthest in)
const MaxZoom = 22

// NewTile returns a Tile of Z,X,Y passed in
func NewTile(z, x, y uint) *Tile {
	return &Tile{
		Z: z,
		X: x,
		Y: y,
	}
}

// Tile describes a slippy tile.
type Tile struct {
	// zoom
	Z uint
	// column
	X uint
	// row
	Y uint
}

// NewTileMinMaxer returns the smallest tile which fits the
// geom.MinMaxer. Note: it assumes the values of ext are
// EPSG:4326 (lng/lat)
//TODO (meilinger): we need this anymore?
func NewTileMinMaxer(ext geom.MinMaxer, tileSRID uint) *Tile {
	upperLeft := NewTileLatLon(MaxZoom, ext.MaxY(), ext.MinX(), tileSRID)
	point := &geom.Point{ext.MaxX(), ext.MinY()}

	var ret *Tile

	for z := uint(MaxZoom); int(z) >= 0 && ret == nil; z-- {
		upperLeft.RangeFamilyAt(z, tileSRID, func(tile *Tile, srid uint) error {
			if tile.Extent4326(tileSRID).Contains(point) {
				ret = tile
				return errors.New("stop iter")
			}

			return nil
		})

	}
	return ret
}

// NewTileLatLon instantiates a tile containing the coordinate with the specified zoom
func NewTileLatLon(z uint, lat, lon float64, srid uint) *Tile {
	grid := GetGrid(srid)
	x := grid.Lon2XIndex(z, lon)
	y := grid.Lat2YIndex(z, lat)

	return &Tile{
		Z: z,
		X: x,
		Y: y,
	}
}

func minmax(a, b uint) (uint, uint) {
	if a > b {
		return b, a
	}
	return a, b
}

// FromBounds returns a list of tiles that make up the bound given. The bounds should be defined as the following lng/lat points [4]float64{west,south,east,north}
func FromBounds(bounds *geom.Extent, z uint, tileSRID uint) []Tile {
	if bounds == nil {
		return nil
	}

	grid := GetGrid(tileSRID)

	minx, maxx := minmax(grid.Lon2XIndex(z, bounds[0]), grid.Lon2XIndex(z, bounds[2]))
	miny, maxy := minmax(grid.Lat2YIndex(z, bounds[1]), grid.Lat2YIndex(z, bounds[3]))

	// tiles := make([]Tile, (maxx-minx)*(maxy-miny))
	var tiles []Tile
	for x := minx; x <= maxx; x++ {
		for y := miny; y <= maxy; y++ {
			tiles = append(tiles, Tile{Z: z, X: x, Y: y})
		}
	}
	return tiles

}

// ZXY returns back the z,x,y of the tile
func (t Tile) ZXY() (uint, uint, uint) { return t.Z, t.X, t.Y }

// Extent gets the extent of the tile in the units of the tileSRID
func (t Tile) NativeExtent(tileSRID uint) *geom.Extent {
	if _, ok := SupportedProjections[tileSRID]; !ok {
		panic(fmt.Sprintf("unsupported tileSRID %v", tileSRID))
	}

	grid := GetGrid(tileSRID)
	pts := []float64{grid.XIndex2Lon(t.Z, t.X), grid.YIndex2Lat(t.Z, t.Y+1), grid.XIndex2Lon(t.Z, t.X+1), grid.YIndex2Lat(t.Z, t.Y)}

	// No need to go further, we've already got the WGS84 extents
	if tileSRID == proj.WGS84 {
		return geom.NewExtent(
			[2]float64{pts[0], pts[1]},
			[2]float64{pts[2], pts[3]},
		)
	}

	pts, err := proj.Convert(proj.EPSGCode(tileSRID), pts)
	if err != nil {
		panic(fmt.Sprintf("error converting %v to %v", pts, tileSRID))
	}

	return geom.NewExtent(
		[2]float64{pts[0], pts[1]},
		[2]float64{pts[2], pts[3]},
	)
}

// Extent3857 returns the tile's extent in EPSG:3857 (aka Web Mercator) projection
func (t Tile) Extent3857(tileSRID uint) *geom.Extent {
	if tileSRID != 3857 {
		// Can't necessarily get webmercator extent for 4326 tile
		panic("unable to get 3857 extent on 4326 tile")
	}
	return geom.NewExtent(
		[2]float64{Tile2WebX(t.Z, t.X, tileSRID), Tile2WebY(t.Z, t.Y+1, tileSRID)},
		[2]float64{Tile2WebX(t.Z, t.X+1, tileSRID), Tile2WebY(t.Z, t.Y, tileSRID)},
	)
}

// Extent4326 returns the tile's extent in EPSG:4326 (aka lat/long) given the tilespace's SRID
func (t Tile) Extent4326(tileSRID uint) *geom.Extent {
	grid := GetGrid(tileSRID)
	return geom.NewExtent(
		[2]float64{grid.XIndex2Lon(t.Z, t.X), grid.YIndex2Lat(t.Z, t.Y+1)},
		[2]float64{grid.XIndex2Lon(t.Z, t.X+1), grid.YIndex2Lat(t.Z, t.Y)},
	)
}

// RangeFamilyAt calls f on every tile vertically related to t at the specified zoom
// TODO (ear7h): sibling support
//TODO (meilinger): should this be part of TileGrid?
func (t Tile) RangeFamilyAt(zoom, srid uint, f func(tile *Tile, srid uint) error) error {
	// handle ancestors and self
	if zoom <= t.Z {
		mag := t.Z - zoom
		arg := NewTile(zoom, t.X>>mag, t.Y>>mag)
		return f(arg, srid)
	}

	// handle descendants
	mag := zoom - t.Z
	delta := uint(math.Exp2(float64(mag)))

	leastX := t.X << mag
	leastY := t.Y << mag

	for x := leastX; x < leastX+delta; x++ {
		for y := leastY; y < leastY+delta; y++ {
			err := f(NewTile(zoom, x, y), srid)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
