package must

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-spatial/geom/encoding/wkt"

	"github.com/go-spatial/geom"
)

var (
	coordRegexText = `-?\d+(?:\.\d+)?`

	pointRegexText      = fmt.Sprintf(`(?P<x>%[1]v)\s+(?P<y>%[1]v)`, coordRegexText)
	singleLineRegexText = fmt.Sprintf(`\(%[1]v(?:\s*,\s*%[1]v)*\)`, pointRegexText)
	multiLineRegexText  = fmt.Sprintf(`,?(%[1]v)`, singleLineRegexText)

	pointRegex      = regexp.MustCompile(pointRegexText)
	singleLineRegex = regexp.MustCompile(singleLineRegexText)
	multiLineRegex  = regexp.MustCompile(multiLineRegexText)
)

// ReadMultilines reads the multiline out of the file
// the lines are expected to follow the multiline format
func ReadMultilines(filename string) []geom.Line {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return ParseMultilines(content)
}

func ParseMultilines(content []byte) []geom.Line {

	geo, err := wkt.DecodeBytes(content)
	if err != nil {
		panic(err)
	}
	ml, ok := geo.(geom.MultiLineString)
	if !ok {
		panic("expected multilinestring")
	}
	lines := make([]geom.Line, len(ml))
	for i := range ml {
		if len(ml[i]) != 2 {
			panic(fmt.Sprintf("line(%v) with more then two points", i))
		}
		lines[i][0] = ml[i][0]
		lines[i][1] = ml[i][1]
	}
	return lines
}

// ReadPoints reads the points out of a file.
// the points are expected to
func ReadPoints(filename string) [][2]float64 {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return PrasePoints(content)

}
func PrasePoints(content []byte) [][2]float64 {
	allIndexes := pointRegex.FindAllSubmatchIndex(content, -1)
	var points [][2]float64

	for _, loc := range allIndexes {
		x, err := strconv.ParseFloat(string(content[loc[2]:loc[3]]), 64)
		if err != nil {
			panic(err)
		}
		y, err := strconv.ParseFloat(string(content[loc[4]:loc[5]]), 64)
		if err != nil {
			panic(err)
		}
		points = append(points, [2]float64{x, y})
	}
	return points
}

func DecodeAsLines(content []byte) (segs []geom.Line) {
	str := strings.NewReader(string(content))
	decoder := wkt.NewDecoder(str)
	g, err := decoder.Decode()
	if err != nil {
		panic(err)
	}
	switch geo := g.(type) {
	case geom.LineString:
		segs, err = geo.AsSegments()
		if err != nil {
			panic(err)
		}
	case geom.MultiLineString:
		s, err := geo.AsSegments()
		if err != nil {
			panic(err)
		}
		for i := range s {
			segs = append(segs, s[i]...)
		}
	case geom.Polygon:
		s, err := geo.AsSegments()
		if err != nil {
			panic(err)
		}
		for i := range s {
			segs = append(segs, s[i]...)
		}
	case geom.MultiPolygon:
		s, err := geo.AsSegments()
		if err != nil {
			panic(err)
		}
		for i := range s {
			for j := range s[i] {
				segs = append(segs, s[i][j]...)
			}
		}
	default:
		panic("geometry not supported for AsLines")
	}
	return segs
}
