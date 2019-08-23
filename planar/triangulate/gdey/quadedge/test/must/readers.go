package must

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"

	"github.com/go-spatial/geom"
)

var (
	coordRegexText = `\d+(?:\.\d+)?`

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
	return lines
}

// ReadPoints reads the points out of a file.
// the points are expected to
func ReadPoints(filename string) [][2]float64 {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
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
