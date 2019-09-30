package testing

import (
	"fmt"
	"github.com/go-spatial/geom"
)

var tilesWkt = []string{_ne_6_43_21, _ne_8_42_98}
var tilesCompiled []geom.Collection

// Tiles returns geom.Collections that represent Natural Earth tiles
// generated with the utility https://github.com/ear7h/tile-dump. The
// tiles must be compiled at runtime by calling the CompileTiles
// function (see documentation for it) or this funciton will panic.
//	package my_test
//
//	import (
//		gtesting "github.com/go-spatial/testing"
//		"github.com/go-spatial/geom/encoding/wkt"
//	)
//
//	func init {
//		// put this in init so benchmarks aren't skewed
//		gtesting.CompileTiles(wkt.DecodeString)
//	}
//
//	func TestMy(t *testing) {
//		tiles := gtesting.Tiles()
//		...
//	}
//
func Tiles() []geom.Collection {
	if tilesCompiled == nil {
		panic("no compiled tiles, make sure to call CompileTiles")
	}

	return tilesCompiled
}

// CompileTiles will compile some Natural Earth tiles so they are accessible
// through the Tiles function. It takes a WKT decoding function which can be
// accessible in the package github.com/go-spatial/geom/encoding/wkt as the
// function DecodeString.
//	package my_test
//
//	import (
//		gtesting "github.com/go-spatial/testing"
//		"github.com/go-spatial/geom/encoding/wkt"
//	)
//
//	func init {
//		// put this in init so benchmarks aren't skewed
//		gtesting.CompileTiles(wkt.DecodeString)
//	}
//
//	func TestMy(t *testing) {
//		tiles := gtesting.Tiles()
//		...
//	}
//
func CompileTiles(wktDecoder func(string) (geom.Geometry, error)) {
	if tilesCompiled != nil {
		return
	}

	ret := make([]geom.Collection, len(tilesWkt))

	for i, v := range tilesWkt {

		col, err := wktDecoder(v)
		if err != nil {
			fmt.Println("i: ", i, "str: ", v[57-10:57+10])
			panic(err)
		}

		ret[i] = col.(geom.Collection)

	}

	tilesCompiled = ret
}
