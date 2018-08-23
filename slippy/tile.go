package slippy

import (
	"math"

	"errors"
	"github.com/go-spatial/geom"
)

const MaxZoom = 22

func NewTile(z, x, y uint) *Tile {
	return &Tile{
		z: z,
		x: x,
		y: y,
	}
}

// Tile describes a slippy tile.
type Tile struct {
	// zoom
	z uint
	// column
	x uint
	// row
	y uint
}

// This function returns the smallest tile which fits the
// geom.MinMaxer. Note: it assumes the values of ext are
// EPSG:4326 (lat/lng)
func NewTileMinMaxer(ext geom.MinMaxer) *Tile {
	upperLeft := NewTileLatLon(MaxZoom, ext.MinX(), ext.MaxY())
	point := &geom.Point{ext.MaxX(), ext.MinY()}

	var ret *Tile

	for z := uint(MaxZoom); z >= 0 && ret == nil; z-- {

		upperLeft.RangeFamilyAt(z, func(tile *Tile) error {
			if tile.Extent4326().Contains(point) {
				ret = tile
				return errors.New("stop iter")
			}

			return nil
		})

	}
	return ret
}

func NewTileLatLon(z uint, lat, lon float64) *Tile {
	x := Lon2Tile(z, lon)
	y := Lat2Tile(z, lat)

	return &Tile{
		z: z,
		x: x,
		y: y,
	}
}

func (t *Tile) ZXY() (uint, uint, uint) { return t.z, t.x, t.y }

// Extent3857 returns the tile's extent in EPSG:3857 (aka Web Mercator) projection
func (t *Tile) Extent3857() *geom.Extent {
	return geom.NewExtent(
		[2]float64{Tile2WebX(t.z, t.x), Tile2WebY(t.z, t.y+1)},
		[2]float64{Tile2WebX(t.z, t.x+1), Tile2WebY(t.z, t.y)},
	)
}

// Extent4326 returns the tile's extent in EPSG:4326 (aka lat/long)
func (t *Tile) Extent4326() *geom.Extent {
	return geom.NewExtent(
		[2]float64{Tile2Lon(t.z, t.x), Tile2Lat(t.z, t.y+1)},
		[2]float64{Tile2Lon(t.z, t.x+1), Tile2Lat(t.z, t.y)},
	)
}

// TODO (ear7h): sibling support
// RangeFamilyAt calls f on every tile vertically related to t at the specified zoom
func (t *Tile) RangeFamilyAt(zoom uint, f func(*Tile) error) error {
	// handle ancestors and self
	if zoom <= t.z {
		mag := t.z - zoom
		arg := NewTile(zoom, t.x>>mag, t.y>>mag)
		return f(arg)
	}

	// handle descendants
	mag := zoom - t.z
	delta := uint(math.Exp2(float64(mag)))

	leastX := t.x << mag
	leastY := t.y << mag

	for x := leastX; x < leastX+delta; x++ {
		for y := leastY; y < leastY+delta; y++ {
			err := f(NewTile(zoom, x, y))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
