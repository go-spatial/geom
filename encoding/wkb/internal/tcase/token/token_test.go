package token

import (
	"reflect"
	"strings"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
)

func TestParseFloat64(t *testing.T) {
	type tcase struct {
		input string
		exp   float64
		err   error
	}
	fn := func(t *testing.T, tc tcase) {
		tt := NewT(strings.NewReader(tc.input))
		f, err := tt.ParseFloat64()
		if tc.err != err {
			t.Errorf(" error, expected %v got %v", tc.err, err)
		}
		if tc.err != nil {
			return
		}
		if tc.exp != f {
			t.Errorf("float64, expected %v got %v", tc.exp, f)
		}
	}
	tests := map[string]tcase{
		"1":  {input: "-12", exp: -12.0},
		"2":  {input: "0", exp: 0.0},
		"3":  {input: "+1_000.00", exp: 1000.0},
		"4":  {input: "-12_000.00", exp: -12000.0},
		"5":  {input: "10.005e5", exp: 10.005e5},
		"6":  {input: "10.005e+5", exp: 10.005e5},
		"7":  {input: "10.005e+05", exp: 10.005e5},
		"8":  {input: "1.0005e+6", exp: 10.005e5},
		"9":  {input: "1.0005e+06", exp: 10.005e5},
		"10": {input: "1.0005e-06", exp: 1.0005e-06},
		"11": {input: "1.0005e-06a", exp: 1.0005e-06},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}
func TestParsePoint(t *testing.T) {
	type tcase struct {
		input string
		exp   [2]float64
		err   error
	}
	fn := func(t *testing.T, tc tcase) {
		tt := NewT(strings.NewReader(tc.input))
		f, err := tt.ParsePoint()
		if tc.err != err {
			t.Errorf("error, expected %v got %v", tc.err, err)
		}
		if tc.exp[0] != f[0] || tc.exp[1] != f[1] {
			t.Errorf(" parse point, expected %v got %v", tc.exp, f)
		}
	}
	tests := map[string]tcase{
		"1": {input: "1,-12", exp: [2]float64{1.0, -12.0}},
		"2": {input: "0 /*x*/ ,0 /*y*/", exp: [2]float64{0.0, 0.0}},
		"3": {input: "  +1_000.00 ,/*y*/ 1", exp: [2]float64{1000.0, 1.0}},
		"4": {input: "-12_000.00,0", exp: [2]float64{-12000.0, 0.0}},
		"5": {input: "/* x */ -12_000.00 /* in dollars */, /* y */ 0 /* ponds */ // This is just for kicks", exp: [2]float64{-12000.0, 0.0}},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}

}
func TestParseMultiPoint(t *testing.T) {
	type tcase struct {
		input string
		exp   geom.MultiPoint
		err   error
	}
	fn := func(t *testing.T, tc tcase) {
		tt := NewT(strings.NewReader(tc.input))
		f, err := tt.ParseMultiPoint()
		if tc.err != err {
			t.Errorf("error, expected %v got %v", tc.err, err)
		}
		if !reflect.DeepEqual(tc.exp, f) {
			t.Errorf("parse multipoint, expected %#v got%#v", tc.exp, f)
		}
	}
	tests := map[string]tcase{
		"1": {
			input: `
			( 1,-12 )
			`,
			exp: geom.MultiPoint([][2]float64{{1.0, -12.0}}),
		},
		"2": {
			input: `
			( 1,-12 0,1)
			`,
			exp: geom.MultiPoint([][2]float64{{1.0, -12.0}, {0.0, 1.0}}),
		},
		"3": {
			input: `
			(1,-12 0,1 1,2 )
			`,
			exp: geom.MultiPoint([][2]float64{{1.0, -12.0}, {0.0, 1.0}, {1.0, 2.0}}),
		},
		"4": {
			input: `
			( 
			1,-12 // Position 1
			/* is x suppose to be this? */ 
			0,1  ///
			1,2_000
			) /* Why the end why? */
			`,
			exp: geom.MultiPoint([][2]float64{{1.0, -12.0}, {0.0, 1.0}, {1.0, 2000.0}}),
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}

func TestParseLineString(t *testing.T) {
	type tcase struct {
		input string
		exp   geom.LineString
		err   error
	}
	fn := func(t *testing.T, tc tcase) {
		tt := NewT(strings.NewReader(tc.input))
		f, err := tt.ParseLineString()
		if tc.err != err {
			t.Errorf("error, expected %v got %v", tc.err, err)
		}
		if !reflect.DeepEqual(tc.exp, f) {
			t.Errorf("parse line string, expected (%#v) %[1]v got (%#v) %[2]v", tc.exp, f)
		}
	}

	tests := map[string]tcase{
		"1": {
			input: `
			[ 1,-12 ]
			`,
			exp: geom.LineString([][2]float64{{1.0, -12.0}}),
		},
		"2": {
			input: `
			[ 1,-12 0,1]
			`,
			exp: geom.LineString([][2]float64{{1.0, -12.0}, {0.0, 1.0}}),
		},
		"3": {
			input: `
			[ 1,-12 0,1 1,2 ]
			`,
			exp: geom.LineString([][2]float64{{1.0, -12.0}, {0.0, 1.0}, {1.0, 2.0}}),
		},
		"4": {
			input: `
			[ 
			1,-12 // Position 1
			/* is x suppose to be this? */ 
			0,1  ///
			1,2_000
			] /* Why the end why? */
			`,
			exp: geom.LineString([][2]float64{{1.0, -12.0}, {0.0, 1.0}, {1.0, 2000.0}}),
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}

func TestParseMultiLineString(t *testing.T) {
	type tcase struct {
		input string
		exp   geom.MultiLineString
		err   error
	}
	fn := func(t *testing.T, tc tcase) {
		tt := NewT(strings.NewReader(tc.input))
		f, err := tt.ParseMultiLineString()
		if tc.err != err {
			t.Errorf("error, expected %v got %v", tc.err, err)
		}
		if !reflect.DeepEqual(tc.exp, f) {
			t.Errorf("parse multi-linestring, expected (%#v) %[1]v got (%#v) %[2]v", tc.exp, f)
		}
	}
	tests := map[string]tcase{
		"1": {
			input: `
			[[
			[ 1,-12 ]
			]]
			`,
			exp: geom.MultiLineString([][][2]float64{{{1.0, -12.0}}}),
		},
		"2": {
			input: `
			[[ [ 1,-12 0,1] ]]
			`,
			exp: geom.MultiLineString([][][2]float64{{{1.0, -12.0}, {0.0, 1.0}}}),
		},
		"3": {
			input: `
			[[ [ 1,-12 0,1 1,2 ] ]]
			`,
			exp: geom.MultiLineString([][][2]float64{{{1.0, -12.0}, {0.0, 1.0}, {1.0, 2.0}}}),
		},
		"4": {
			input: `
			[[ [ 1,-12 0,1 1,2 ] [ 1, 2] ]]
			`,
			exp: geom.MultiLineString([][][2]float64{{{1.0, -12.0}, {0.0, 1.0}, {1.0, 2.0}}, {{1.0, 2.0}}}),
		},
		"5": {
			input: `[[
			[ 
			1,-12 // Position 1
			/* is x suppose to be this? */ 
			0,1  ///
			1,2_000
			] /* Why the end why? */
			]]`,
			exp: geom.MultiLineString([][][2]float64{{{1.0, -12.0}, {0.0, 1.0}, {1.0, 2000.0}}}),
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}
func TestParsePolygon(t *testing.T) {
	type tcase struct {
		input string
		exp   geom.Polygon
		err   error
	}
	fn := func(t *testing.T, tc tcase) {
		tt := NewT(strings.NewReader(tc.input))
		f, err := tt.ParsePolygon()
		if tc.err != err {
			t.Errorf("error, expected %v got %v ", tc.err, err)
		}
		if !reflect.DeepEqual(tc.exp, f) {
			t.Errorf("parse polygon, expected (%#v) %[1]v got (%#v) %[2]v", tc.exp, f)
		}
	}
	tests := map[string]tcase{
		"1": {
			input: `
			{
			[ 1,-12 ]
			}
			`,
			exp: geom.Polygon([][][2]float64{{{1.0, -12.0}}}),
		},
		"2": {
			input: `
			{ [ 1,-12 0,1]
		}
			`,
			exp: geom.Polygon([][][2]float64{{{1.0, -12.0}, {0.0, 1.0}}}),
		},
		"3": {
			input: `
			{ [ 1,-12 0,1 1,2 ] }`,
			exp: geom.Polygon([][][2]float64{{{1.0, -12.0}, {0.0, 1.0}, {1.0, 2.0}}}),
		},
		"4": {
			input: `
			{ [ 1,-12 0,1 1,2 ] [ 1, 2]} 
			`,
			exp: geom.Polygon([][][2]float64{{{1.0, -12.0}, {0.0, 1.0}, {1.0, 2.0}}, {{1.0, 2.0}}}),
		},
		"5": {
			input: `{
			[ 
			1,-12 // Position 1
			/* is x suppose to be this? */ 
			0,1  ///
			1,2_000
			] /* Why the end why? */
		}`,
			exp: geom.Polygon([][][2]float64{{{1.0, -12.0}, {0.0, 1.0}, {1.0, 2000.0}}}),
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}
func TestParseMultiPolygon(t *testing.T) {
	type tcase struct {
		input string
		exp   geom.MultiPolygon
		err   error
	}
	fn := func(t *testing.T, tc tcase) {
		tt := NewT(strings.NewReader(tc.input))
		f, err := tt.ParseMultiPolygon()
		if tc.err != err {
			t.Errorf("error, expected: %v got %v", tc.err, err)
		}
		if !reflect.DeepEqual(tc.exp, f) {
			t.Errorf("parse multi-polygon expected (%#v) %[1]v got (%#v) %[2]v", tc.exp, f)
		}
	}
	tests := map[string]tcase{
		"1": {
			input: `{{
			{
			[ 1,-12 ]
			}
		}}`,
			exp: geom.MultiPolygon([][][][2]float64{{{{1.0, -12.0}}}}),
		},
		"2": {
			input: `{{
			{ [ 1,-12 0,1]
		} }}
			`,
			exp: geom.MultiPolygon([][][][2]float64{{{{1.0, -12.0}, {0.0, 1.0}}}}),
		},
		"3": {
			input: `
			{{ { [ 1,-12 0,1 1,2 ] } }}`,
			exp: geom.MultiPolygon([][][][2]float64{{{{1.0, -12.0}, {0.0, 1.0}, {1.0, 2.0}}}}),
		},
		"4": {
			input: `
			{{{ [ 1,-12 0,1 1,2 ] [ 1, 2]} 
		}}`,
			exp: geom.MultiPolygon([][][][2]float64{{{{1.0, -12.0}, {0.0, 1.0}, {1.0, 2.0}}, {{1.0, 2.0}}}}),
		},
		"5": {
			input: `{{{
			[ 
			1,-12 // Position 1
			/* is x suppose to be this? */ 
			0,1  ///
			1,2_000
			] /* Why the end why? */}
		}}`,
			exp: geom.MultiPolygon([][][][2]float64{{{{1.0, -12.0}, {0.0, 1.0}, {1.0, 2000.0}}}}),
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}

func TestParseCollection(t *testing.T) {
	type tcase struct {
		input string
		exp   geom.Collection
		err   error
	}
	fn := func(t *testing.T, tc tcase) {
		tt := NewT(strings.NewReader(tc.input))
		f, err := tt.ParseCollection()
		if tc.err != err {
			t.Errorf("error, expected %v got %v", tc.err, err)
		}
		if !cmp.CollectionerEqual(tc.exp, f) {
			t.Errorf("parse collection, \nexpected (%#v) %[1]v \ngot      (%#v) %[2]v", tc.exp, f)
		}
	}
	tests := map[string]tcase{
		"1": {
			input: `(( {{
			{
			[ 1,-12 ]
			}
		}}))`,
			exp: geom.Collection{
				geom.MultiPolygon([][][][2]float64{{{{1.0, -12.0}}}}),
			},
		},
		"2": {
			input: `(( {{
			{ [ 1,-12 0,1]
		} }}
	))`,
			exp: geom.Collection{
				geom.MultiPolygon([][][][2]float64{{{{1.0, -12.0}, {0.0, 1.0}}}}),
			},
		},
		"3": {
			input: `((
			{{ { [ 1,-12 0,1 1,2 ] } }} ))`,
			exp: geom.Collection{
				geom.MultiPolygon([][][][2]float64{{{{1.0, -12.0}, {0.0, 1.0}, {1.0, 2.0}}}}),
			},
		},
		"4": {
			input: `((
			{{{ [ 1,-12 0,1 1,2 ] [ 1, 2]} 
		}}))`,
			exp: geom.Collection{
				geom.MultiPolygon([][][][2]float64{{{{1.0, -12.0}, {0.0, 1.0}, {1.0, 2.0}}, {{1.0, 2.0}}}}),
			},
		},
		"5": {
			input: `(( {{{
			[ 
			1,-12 // Position 1
			/* is x suppose to be this? */ 
			0,1  ///
			1,2_000
			] /* Why the end why? */}
		}} ))`,
			exp: geom.Collection{
				geom.MultiPolygon([][][][2]float64{{{{1.0, -12.0}, {0.0, 1.0}, {1.0, 2000.0}}}}),
			},
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}

func TestParseBinary(t *testing.T) {
	type tcase struct {
		input string
		exp   []byte
		err   error
	}
	fn := func(t *testing.T, tc tcase) {
		tt := NewT(strings.NewReader(tc.input))
		c, err := tt.ParseBinaryField()
		if tc.err != err {
			t.Errorf("error, expected %v got %v", tc.err, err)
		}
		if !reflect.DeepEqual(tc.exp, c) {
			t.Errorf("parse binary, expected %v got %v", tc.exp, c)
		}
	}
	tests := map[string]tcase{
		"simple": {
			input: `
// 	01 02 03 04  05 06 07 08
{{
	01                       // Byte order Marker little
	02 00 00 00              // Type 2 LineString
	02 00 00 00              // number of points
	00 00 00 00  00 00 F0 3F // x 1
	00 00 00 00  00 00 00 40 // y 2
	00 00 00 00  00 00 08 40 // x 3
	00 00 00 00  00 00 10 40 // y 4
}}`,
			exp: []byte{
				0x01,
				0x02, 0x00, 0x00, 0x00,
				0x02, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0x3f,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x08, 0x40,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x40,
			},
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}

func TestParseLabel(t *testing.T) {
	type tcase struct {
		input string
		exp   string
		err   error
	}
	fn := func(t *testing.T, tc tcase) {
		tt := NewT(strings.NewReader(tc.input))
		c, err := tt.ParseLabel()
		if tc.err != err {
			t.Errorf("error, expected %v got %v", tc.err, err)
		}
		if !reflect.DeepEqual(tc.exp, string(c)) {
			t.Errorf("parse label, expected %v got %v", tc.exp, string(c))
		}
	}
	tests := map[string]tcase{
		"easy": {
			input: "easy: an easy label",
			exp:   "easy",
		},
		"little-easy": {
			input: "little-easy: an easy label",
			exp:   "little-easy",
		},
		"little_easy": {
			input: `
			
			
			
			little_easy: an easy label
			
			
			`,
			exp: "little_easy",
		},
		"little.easy": {
			input: `
			/*
			    This one is also pretty easy.
			*/
			
			little.easy: an easy label
			
			
			`,
			exp: "little.easy",
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}

func TestPraseLineComment(t *testing.T) {
	type tcase struct {
		exp      string
		dontwrap bool
		err      error
	}
	fn := func(t *testing.T, tc tcase) {
		input := tc.exp
		if !tc.dontwrap {
			input = "//" + input + "\n"
		}
		tt := NewT(strings.NewReader(input))
		c, err := tt.ParseLineComment()
		if tc.err != err {
			t.Errorf("error, expected %v got %v", tc.err, err)
		}
		if !reflect.DeepEqual(tc.exp, string(c)) {
			t.Errorf("prase comment, expected “%v” got “%v”", tc.exp, string(c))
		}
	}
	tests := map[string]tcase{
		"simple": {
			exp: `This is a string. `,
		},
		"slashes inside": {
			exp: " this comment // and another.",
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}

func TestParseComment(t *testing.T) {

	type tcase struct {
		exp      string
		dontwrap bool
		err      error
	}

	fn := func(t *testing.T, tc tcase) {
		input := tc.exp
		if !tc.dontwrap {
			input = "/*" + input + "*/"
		}
		tt := NewT(strings.NewReader(input))
		c, err := tt.ParseComment()
		if tc.err != err {
			t.Errorf("error, expected %v got %v", tc.err, err)
		}
		if !reflect.DeepEqual(tc.exp, string(c)) {
			t.Errorf("parsing comment, expected %v got %v", tc.exp, string(c))
		}
	}
	tests := map[string]tcase{
		"simple": {
			exp: `This is a string. 
			In a multiline comment.`,
		},
		"empty": {
			exp: "",
		},
		"empty with inline": {
			exp: "//",
		},
		"spread out empty": {
			exp: `




			/**/

			`,
		},
		"complex": {
			exp: `
			This is a line // with an line comment that does not mean anything.

			/******************8
			more stuff
			*************.
			* /
			*/
			`,
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}

}
