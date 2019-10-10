package slippy

import (
	"fmt"
	"math"

	"github.com/go-spatial/geom"
)

type extents struct {
	NativeExtents *geom.Extent
	WGS84Extents  *geom.Extent
	Grid          TileGrid
}

// SupportedProjections contains supported projection native and lat/long extents as well as tile layout ratio
var SupportedProjections = map[uint]extents{
	3857: extents{NativeExtents: &geom.Extent{-20026376.39, -20048966.10, 20026376.39, 20048966.10}, WGS84Extents: &geom.Extent{-180.0, -85.0511, 180.0, 85.0511}, Grid: GetGrid(3857)},
	4326: extents{NativeExtents: &geom.Extent{-180.0, -90.0, 180.0, 90.0}, WGS84Extents: &geom.Extent{-180.0, -90.0, 180.0, 90.0}, Grid: GetGrid(4326)},
}

// ==== Web Mercator ====

// WebMercatorMax is the max size in meters of a tile
const WebMercatorMax = 20037508.34

// Tile2WebX returns the side of the tile in the -x side in webmercator
func Tile2WebX(zoom uint, n uint, srid uint) float64 {
	res := (WebMercatorMax * 2) / math.Exp2(float64(zoom))
	return -WebMercatorMax + float64(n)*res
}

// Tile2WebY returns the side of the tile in the +y side in webmercator
func Tile2WebY(zoom uint, n uint, srid uint) float64 {
	res := (WebMercatorMax * 2) / math.Exp2(float64(zoom))

	return WebMercatorMax - float64(n)*res
}

// ==== pixels ====

// MvtTileDim is the number of pixels in a tile
const MvtTileDim = 4096.0

// PixelsToProjectedUnits scalar conversion of pixels into projected units
// TODO (@ear7h): perhaps rethink this
func PixelsToProjectedUnits(zoom uint, pixels uint, srid uint) float64 {
	switch srid {
	case 3857:
		return WebMercatorMax * 2 / math.Exp2(float64(zoom)) * float64(pixels) / MvtTileDim
	case 4326:
		return 360.0 / math.Exp2(float64(zoom)) * float64(pixels) / MvtTileDim / 2
	default:
		panic(fmt.Sprintf("unsupported srid: %v", srid))
	}
}
