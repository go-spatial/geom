package makevalid

import (
	"fmt"
	"strings"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/wkt"
)

/*
The purpose of this file is to be a holding ground
for usefull function that would otherwise clutter up
the test case. Generally these should be functions
that are going to be wrapped in debug blocks and
dump to the terminal helpful debugging information.


The general format for these functions should be the
following:

func dumpDescriptiveName(...Paramaters...) string
*/


// dumpWKTLineSegments will dump out both provided lineSegments next to each other.
func dumpWKTLineSegments(format string, lineSegment1, lineSegment2 []geom.Line) string {

	var b strings.Builder

	if format == "" {
		format = "%04d : %s | %s\n"
	}
	l1Larger := true
	maxi := len(lineSegment1)
	smax := len(lineSegment2)
	if len(lineSegment2) > maxi {
		maxi, smax = smax, maxi
		l1Larger = false
	}

	for i := 0; i < maxi; i++ {

		v1, v2 := "\t", "\t"
		if i < smax {
			v1, v2 = wkt.MustEncode(lineSegment1[i]), wkt.MustEncode(lineSegment2[i])
		} else if l1Larger {
			v1 = wkt.MustEncode(lineSegment1[i])
		} else {
			v2 = wkt.MustEncode(lineSegment2[i])
		}
		fmt.Fprintf(&b, format, i, v1, v2)
	}
	return b.String()
}

// dumpWKTLineSegment will dump the given lineSegments encodeing each lineSegment as WKT
func dumpWKTLineSegment(title, format string, lineSegments []geom.Line) string {
	var b strings.Builder

	if format == "" {
		format = "%04d : %s\n"
	}

	b.WriteString(title)
	b.WriteString("\n")

	for i := 0; i < len(lineSegments); i++ {
		fmt.Fprintf(&b, format, i, wkt.MustEncode(lineSegments[i]))
	}
	return b.String()
}
