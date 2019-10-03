package slippy

import (
	"fmt"
	"math"

	"github.com/go-spatial/geom"
)

type extents struct {
	NativeExtents *geom.Extent
	WGS84Extents  *geom.Extent
	W2HTileRatio  float64
}

// SupportedProjectionExtents contains supported projection native and lat/long extents as well as tile layout ratio
var SupportedProjections = map[uint]extents{
	3857: extents{NativeExtents: &geom.Extent{-20026376.39, -20048966.10, 20026376.39, 20048966.10}, WGS84Extents: &geom.Extent{-180.0, -85.0511, 180.0, 85.0511}, W2HTileRatio: 1.0},
	4326: extents{NativeExtents: &geom.Extent{-180.0, -90.0, 180.0, 90.0}, WGS84Extents: &geom.Extent{-180.0, -90.0, 180.0, 90.0}, W2HTileRatio: 2.0},
}

// Lat2Tile takes a zoom, lat and tilespace srid to produce the tile y index
func Lat2Tile(zoom uint, lat float64, srid uint) (y uint) {
	switch srid {
	case 3857:
		latRad := lat * math.Pi / 180
		return uint(math.Exp2(float64(zoom))*
			(1.0-math.Log(
				math.Tan(latRad)+
					(1/math.Cos(latRad)))/math.Pi)) /
			2.0
	case 4326:
		ratio := SupportedProjections[srid].W2HTileRatio
		return uint(math.Exp2(float64(zoom)) * -(lat - 90.0) / (360.0 / ratio))
	default:
		panic(fmt.Sprintf("unsupported srid: %v", srid))
	}
}

// Lon2Tile takes in a zoom, lat and tilespace srid to produce the tile x index
func Lon2Tile(zoom uint, lon float64, srid uint) (x uint) {
	switch srid {
	case 3857:
		fallthrough
	case 4326:
		ratio := SupportedProjections[srid].W2HTileRatio
		return uint(math.Exp2(float64(zoom)) * (lon + 180.0) / (360.0 / ratio))
	default:
		panic(fmt.Sprintf("unsupported srid: %v", srid))
	}
}

// Tile2Lon will return the west most longitude
func Tile2Lon(zoom, x uint, srid uint) float64 {
	switch srid {
	case 3857:
		fallthrough
	case 4326:
		ratio := SupportedProjections[srid].W2HTileRatio
		return float64(x)/math.Exp2(float64(zoom))*(360.0/ratio) - 180.0
	default:
		panic(fmt.Sprintf("unsupported srid: %v", srid))
	}
}

// Tile2Lat will return the north most latitude
func Tile2Lat(zoom, y uint, srid uint) float64 {
	switch srid {
	case 3857:
		var n = math.Pi
		if y != 0 {
			n = math.Pi - 2.0*math.Pi*float64(y)/math.Exp2(float64(zoom))
		}

		return 180.0 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
	case 4326:
		return -(180.0/math.Exp2(float64(zoom))*float64(y) - 90.0)
	default:
		panic(fmt.Sprintf("unsupported srid: %v", srid))
	}
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
