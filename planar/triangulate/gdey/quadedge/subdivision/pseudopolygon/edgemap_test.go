package pseudopolygon

import (
	"fmt"
	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"testing"
)

func TestNormalizeLine(t *testing.T){
	type tcase struct {
		Desc string
		line geom.Line
		expected geom.Line
	}

	fn := func(tc tcase) func(*testing.T){
		return func(t *testing.T){

			got := tc.line
			normalizeLine(&got)
			pt1 := cmp.GeomPointEqual(got[0],tc.expected[0])
			pt2 := cmp.GeomPointEqual(got[1],tc.expected[1])
			if !( pt1 && pt2 ){
				t.Errorf("line, expected %v got %v",tc.expected,got)
				t.Logf("pt1: %v -- %v , %v",pt1,tc.expected[0],got[0])
				t.Logf("pt2: %v -- %v , %v",pt2,tc.expected[1],got[1])
			}

		}
	}

	tests := []tcase{
		//subtests
		{
			line: geom.Line{{0,0},{1,1}},
			expected: geom.Line{{1,1},{0,0}},
		},
		{
			line: geom.Line{{1,1},{0,0}},
			expected: geom.Line{{1,1},{0,0}},
		},
		{
			line: geom.Line{{1,1},{1,1}},
			expected: geom.Line{{1,1},{1,1}},
		},
	}
	for i := range tests {
		t.Run(tests[i].Desc, fn(tests[i]))
	}
}

func TestEdges(t *testing.T) {
	type tcase struct {
		Desc string
		points []geom.Point
		lines []geom.Line
	}

	fn := func(tc tcase) func(*testing.T){
		return func(t *testing.T){

			lnsmap := make(map[geom.Line]bool,len(tc.lines))
			for i := range tc.lines {
				normalizeLine(&(tc.lines[i]))
				lnsmap[tc.lines[i]] = true
			}

			em := newEdgeMap(tc.points)
			if em == nil {
				t.Errorf("edgemap, expected not nil, got nil")
				return
			}

			lines := em.Edges()

			if len(lines) != len(tc.lines) {
				t.Errorf("number of lines, expected %v got %v",len(tc.lines),len(lines))

				ln := len(tc.lines)
				if len(lines) > ln {
					ln = len(lines)
				}

				for i:= 0; i < ln; i++ {
					 tcstr, lnstr := "-----", "-----"
					 if i < len(tc.lines) {
					 	tcstr = fmt.Sprintf("%v",tc.lines[i])
					 }
					if i < len(lines) {
						lnstr = fmt.Sprintf("%v",lines[i])
					}
					 t.Logf("\t%03v : %v : %v ",i,tcstr,lnstr)

				}

				return
			}

			for i := range tc.lines {
				pt1 := cmp.GeomPointEqual(lines[i][0],tc.lines[i][0])
				pt2 := cmp.GeomPointEqual(lines[i][1],tc.lines[i][1])
				if !( pt1 && pt2 ){
					t.Errorf("line %v, expected %v got %v",i,tc.lines[i],lines[i])
				}
			}

		}
	}

	tests := []tcase{
		//subtests
		{
			points: []geom.Point{
				{0,-2},{3,-2},{4,-1},{4,1},{3,2},{0,2},{-3,2},{4,1},{-3,-2},
			},
			lines: []geom.Line{
				{{4,1}, {-3,-2}},
				{{-3,2}, {4,1}},
				{{3,2}, {0,2}},
				{{0,-2},{3,-2}},
				{{0,2}, {-3,2}},
				{{-3,-2},{0,-2}},
				{{4,-1},{4,1}},
				{{4,1}, {3,2}},
				{{3,-2},{4,-1}},
			},
		},

	}
	for i := range tests {
		t.Run(tests[i].Desc, fn(tests[i]))
	}

}
