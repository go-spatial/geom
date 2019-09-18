package must

import (
	"fmt"
	"github.com/go-spatial/geom/encoding/wkt"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"

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

	allIndexes := singleLineRegex.FindAllSubmatchIndex(content, -1)
	var lines []geom.Line

	for _, loc := range allIndexes {

		x1, err := strconv.ParseFloat(string(content[loc[2]:loc[3]]), 64)
		if err != nil {
			panic(err)
		}
		y1, err := strconv.ParseFloat(string(content[loc[4]:loc[5]]), 64)
		if err != nil {
			panic(err)
		}

		x2, err := strconv.ParseFloat(string(content[loc[6]:loc[7]]), 64)
		if err != nil {
			panic(err)
		}

		y2, err := strconv.ParseFloat(string(content[loc[8]:loc[9]]), 64)
		if err != nil {
			panic(err)
		}

		lines = append(lines, geom.Line{{x1, y1}, {x2, y2}})
	}
	log.Printf("got %v lines:\n%v", len(lines), allIndexes)

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
			segs = append(segs,s[i]...)
		}
	case geom.Polygon:
		s, err := geo.AsSegments()
		if err != nil {
			panic(err)
		}
		for i := range s {
				segs = append(segs,s[i]...)
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

