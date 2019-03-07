package recorder

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/go-spatial/geom/encoding/wkt"
)

type TestDescription struct {
	Name string
	Category string
	Description string
}

type Interface interface {
	Close() error
	Record(geom interface{}, FFL FuncFileLineType, Description TestDescription) error
}

func TypeAndWKT(geom interface{}) (string, string) {
	wktGeom := wkt.MustEncode(geom)
	switch {
	case strings.HasPrefix(wktGeom, "POINT"):
		return "point", wktGeom
	case strings.HasPrefix(wktGeom, "MULTIPOINT"):
		return "multipoint", wktGeom
	case strings.HasPrefix(wktGeom, "LINESTRING"):
		return "linestring", wktGeom
	case strings.HasPrefix(wktGeom, "MULTILINESTRING"):
		return "multilinestring", wktGeom
	case strings.HasPrefix(wktGeom, "POLYGON"):
		return "polygon", wktGeom
	case strings.HasPrefix(wktGeom, "MULTIPOLYGON"):
		return "multipolygon", wktGeom
	default:
		panic(fmt.Sprintf("Unknown wkt type: %v", wktGeom))

	}
	return "", ""
}

type FuncFileLineType struct {
	Func       string
	File       string
	LineNumber int
}

// FuncFileLine returns the func file and line number of the the number of callers
// above the caller of this function. Zero returns the immediate caller above the
// caller of the FuncFileLine func.
func FuncFileLine(lvl uint) FuncFileLineType {
	fnName := "unknown"
	file := "unknown"
	lineNo := -1

	pc, _, _, ok := runtime.Caller(int(lvl) + 2)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		fnName = details.Name()
		file, lineNo = details.FileLine(pc)
	}

	vs := strings.Split(fnName, "/")
	fnName = vs[len(vs)-1]

	return FuncFileLineType{
		Func:       fnName,
		File:       file,
		LineNumber: lineNo,
	}
}
