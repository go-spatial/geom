/*
Copyright (c) 2016 Vivid Solutions.

All rights reserved. This program and the accompanying materials
are made available under the terms of the Eclipse Public License v1.0
and Eclipse Distribution License v. 1.0 which accompanies this distribution.
The Eclipse Public License is available at http://www.eclipse.org/legal/epl-v10.html
and the Eclipse Distribution License is available at

http://www.eclipse.org/org/documents/edl-v10.php.
*/

package triangulate

import (
	"encoding/hex"
	"strconv"
	"testing"

	"github.com/go-spatial/geom/encoding/wkb"
	"github.com/go-spatial/geom/encoding/wkt"
)

/*
TestDelaunayTriangulation test cases were taken from JTS and converted to
GeoJSON.
*/
func TestDelaunayTriangulation(t *testing.T) {
	type tcase struct {
		// provided for readability
		inputWKT string
		// this can be removed if/when geom has a WKT decoder.
		// A simple website for performing conversions:
		// https://rodic.fr/blog/online-conversion-between-geometric-formats/
		inputWKB      string
		expectedEdges string
		expectedTris  string
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			bytes, err := hex.DecodeString(tc.inputWKB)
			if err != nil {
				t.Fatalf("error decoding hex string: %v", err)
				return
			}
			sites, err := wkb.DecodeBytes(bytes)
			if err != nil {
				t.Fatalf("error decoding WKB: %v", err)
				return
			}

			builder := NewDelaunayTriangulationBuilder(1e-6)
			builder.SetSites(sites)
			if builder.create() == false {
				t.Errorf("error building triangulation, expected true got false")
			}
			err = builder.subdiv.Validate()
			if err != nil {
				t.Errorf("error, expected nil got %v", err)
			}

			edges := builder.GetEdges()
			edgesWKT, err := wkt.EncodeString(edges)
			if err != nil {
				t.Errorf("error, expected nil got %v", err)
				return
			}
			if edgesWKT != tc.expectedEdges {
				t.Errorf("error, expected %v got %v", tc.expectedEdges, edgesWKT)
				return
			}

			tris, err := builder.GetTriangles()
			if err != nil {
				t.Errorf("error, expected nil got %v", err)
				return
			}
			trisWKT, err := wkt.EncodeString(tris)
			if err != nil {
				t.Errorf("error, expected nil got %v", err)
				return
			}
			if trisWKT != tc.expectedTris {
				t.Errorf("error, expected %v got %v", tc.expectedTris, trisWKT)
				return
			}
		}
	}
	testcases := []tcase{
		{
			inputWKT:      `MULTIPOINT (10 10, 10 20, 20 20)`,
			inputWKB:      `010500000003000000010200000002000000000000000000244000000000000034400000000000003440000000000000344001020000000200000000000000000024400000000000002440000000000000244000000000000034400102000000020000000000000000002440000000000000244000000000000034400000000000003440`,
			expectedEdges: `MULTILINESTRING ((10 20,20 20),(10 10,10 20),(10 10,20 20))`,
			// the ordering and values are the same as JTS, but reformatted as a
			// MULTIPOLYGON
			expectedTris: `MULTIPOLYGON (((10 20,10 10,20 20,10 20)))`,
		},
		{
			inputWKT: "MULTIPOINT (10 10, 10 20, 20 20, 20 10, 20 0, 10 0, 0 0, 0 10, 0 20)",
			inputWKB: "010400000009000000010100000000000000000024400000000000002440010100000000000000000024400000000000003440010100000000000000000034400000000000003440010100000000000000000034400000000000002440010100000000000000000034400000000000000000010100000000000000000024400000000000000000010100000000000000000000000000000000000000010100000000000000000000000000000000002440010100000000000000000000000000000000003440",
			// This is not the same ordering as JTS, but the segments are the same.
			// This ordering appears to be correct. JTS uses these as primary edges:
			// (10 10,0 20),(10 10,0 10)
			// Where according to the rules in getPrimary, it appears that these
			// should be the primary edges:
			// (0 20,10 10),(0 10,10 10)
			// Since this appears to be more correct, I will not try to make them
			// consistent.
			expectedEdges: "MULTILINESTRING ((10 20,20 20),(0 20,10 20),(0 10,0 20),(0 0,0 10),(0 0,10 0),(10 0,20 0),(20 0,20 10),(20 10,20 20),(10 20,20 10),(10 10,20 10),(10 10,10 20),(0 20,10 10),(0 10,10 10),(10 0,10 10),(0 10,10 0),(10 10,20 0))",
			// the ordering and values are the same as JTS, but reformatted as a
			// MULTIPOLYGON
			expectedTris: "MULTIPOLYGON (((0 20,0 10,10 10,0 20)),((0 20,10 10,10 20,0 20)),((10 20,10 10,20 10,10 20)),((10 20,20 10,20 20,10 20)),((10 0,20 0,10 10,10 0)),((10 0,10 10,0 10,10 0)),((10 0,0 10,0 0,10 0)),((10 10,20 0,20 10,10 10)))",
		},
		{
			inputWKT:      "MULTIPOINT ((50 40),(140 70),(80 100),(130 140),(30 150),(70 180),(190 110),(120 20))",
			inputWKB:      "01040000000800000001010000000000000000004940000000000000444001010000000000000000806140000000000080514001010000000000000000005440000000000000594001010000000000000000406040000000000080614001010000000000000000003e400000000000c0624001010000000000000000805140000000000080664001010000000000000000c067400000000000805b4001010000000000000000005e400000000000003440",
			expectedEdges: "MULTILINESTRING ((70 180,190 110),(30 150,70 180),(30 150,50 40),(50 40,120 20),(120 20,190 110),(120 20,140 70),(140 70,190 110),(130 140,140 70),(130 140,190 110),(70 180,130 140),(80 100,130 140),(70 180,80 100),(30 150,80 100),(50 40,80 100),(80 100,120 20),(80 100,140 70))",
			expectedTris:  "MULTIPOLYGON (((30 150,50 40,80 100,30 150)),((30 150,80 100,70 180,30 150)),((70 180,80 100,130 140,70 180)),((70 180,130 140,190 110,70 180)),((190 110,130 140,140 70,190 110)),((190 110,140 70,120 20,190 110)),((120 20,140 70,80 100,120 20)),((120 20,80 100,50 40,120 20)),((80 100,140 70,130 140,80 100)))",
		},
		{
			inputWKT:      "POLYGON ((42 30, 41.96 29.61, 41.85 29.23, 41.66 28.89, 41.41 28.59, 41.11 28.34, 40.77 28.15, 40.39 28.04, 40 28, 39.61 28.04, 39.23 28.15, 38.89 28.34, 38.59 28.59, 38.34 28.89, 38.15 29.23, 38.04 29.61, 38 30, 38.04 30.39, 38.15 30.77, 38.34 31.11, 38.59 31.41, 38.89 31.66, 39.23 31.85, 39.61 31.96, 40 32, 40.39 31.96, 40.77 31.85, 41.11 31.66, 41.41 31.41, 41.66 31.11, 41.85 30.77, 41.96 30.39, 42 30))",
			inputWKB:      "0103000000010000002100000000000000000045400000000000003e407b14ae47e1fa44405c8fc2f5289c3d40cdccccccccec44407b14ae47e13a3d4014ae47e17ad44440a4703d0ad7e33c4014ae47e17ab44440d7a3703d0a973c40ae47e17a148e4440d7a3703d0a573c40c3f5285c8f6244406666666666263c4052b81e85eb3144400ad7a3703d0a3c4000000000000044400000000000003c40ae47e17a14ce43400ad7a3703d0a3c403d0ad7a3709d43406666666666263c4052b81e85eb714340d7a3703d0a573c40ec51b81e854b4340d7a3703d0a973c40ec51b81e852b4340a4703d0ad7e33c4033333333331343407b14ae47e13a3d4085eb51b81e0543405c8fc2f5289c3d4000000000000043400000000000003e4085eb51b81e054340a4703d0ad7633e40333333333313434085eb51b81ec53e40ec51b81e852b43405c8fc2f5281c3f40ec51b81e854b4340295c8fc2f5683f4052b81e85eb714340295c8fc2f5a83f403d0ad7a3709d43409a99999999d93f40ae47e17a14ce4340f6285c8fc2f53f400000000000004440000000000000404052b81e85eb314440f6285c8fc2f53f40c3f5285c8f6244409a99999999d93f40ae47e17a148e4440295c8fc2f5a83f4014ae47e17ab44440295c8fc2f5683f4014ae47e17ad444405c8fc2f5281c3f40cdccccccccec444085eb51b81ec53e407b14ae47e1fa4440a4703d0ad7633e4000000000000045400000000000003e40",
			expectedEdges: "MULTILINESTRING ((41.66 31.11,41.85 30.77),(41.41 31.41,41.66 31.11),(41.11 31.66,41.41 31.41),(40.77 31.85,41.11 31.66),(40.39 31.96,40.77 31.85),(40 32,40.39 31.96),(39.61 31.96,40 32),(39.23 31.85,39.61 31.96),(38.89 31.66,39.23 31.85),(38.59 31.41,38.89 31.66),(38.34 31.11,38.59 31.41),(38.15 30.77,38.34 31.11),(38.04 30.39,38.15 30.77),(38 30,38.04 30.39),(38 30,38.04 29.61),(38.04 29.61,38.15 29.23),(38.15 29.23,38.34 28.89),(38.34 28.89,38.59 28.59),(38.59 28.59,38.89 28.34),(38.89 28.34,39.23 28.15),(39.23 28.15,39.61 28.04),(39.61 28.04,40 28),(40 28,40.39 28.04),(40.39 28.04,40.77 28.15),(40.77 28.15,41.11 28.34),(41.11 28.34,41.41 28.59),(41.41 28.59,41.66 28.89),(41.66 28.89,41.85 29.23),(41.85 29.23,41.96 29.61),(41.96 29.61,42 30),(41.96 30.39,42 30),(41.85 30.77,41.96 30.39),(41.66 31.11,41.96 30.39),(41.41 31.41,41.96 30.39),(41.41 28.59,41.96 30.39),(41.41 28.59,41.41 31.41),(38.59 28.59,41.41 28.59),(38.59 28.59,41.41 31.41),(38.59 28.59,38.59 31.41),(38.59 31.41,41.41 31.41),(38.59 31.41,39.61 31.96),(39.61 31.96,41.41 31.41),(39.61 31.96,40.39 31.96),(40.39 31.96,41.41 31.41),(40.39 31.96,41.11 31.66),(38.04 30.39,38.59 28.59),(38.04 30.39,38.59 31.41),(38.04 30.39,38.34 31.11),(38.04 29.61,38.59 28.59),(38.04 29.61,38.04 30.39),(39.61 28.04,41.41 28.59),(38.59 28.59,39.61 28.04),(38.89 28.34,39.61 28.04),(40.39 28.04,41.41 28.59),(39.61 28.04,40.39 28.04),(41.96 29.61,41.96 30.39),(41.41 28.59,41.96 29.61),(41.66 28.89,41.96 29.61),(40.39 28.04,41.11 28.34),(38.04 29.61,38.34 28.89),(38.89 31.66,39.61 31.96))",
			expectedTris:  "MULTIPOLYGON (((38.15 30.77,38.04 30.39,38.34 31.11,38.15 30.77)),((38.34 31.11,38.04 30.39,38.59 31.41,38.34 31.11)),((38.59 31.41,38.04 30.39,38.59 28.59,38.59 31.41)),((38.59 31.41,38.59 28.59,41.41 31.41,38.59 31.41)),((38.59 31.41,41.41 31.41,39.61 31.96,38.59 31.41)),((38.59 31.41,39.61 31.96,38.89 31.66,38.59 31.41)),((38.89 31.66,39.61 31.96,39.23 31.85,38.89 31.66)),((39.61 31.96,41.41 31.41,40.39 31.96,39.61 31.96)),((39.61 31.96,40.39 31.96,40 32,39.61 31.96)),((40.39 31.96,41.41 31.41,41.11 31.66,40.39 31.96)),((40.39 31.96,41.11 31.66,40.77 31.85,40.39 31.96)),((41.41 31.41,38.59 28.59,41.41 28.59,41.41 31.41)),((41.41 31.41,41.41 28.59,41.96 30.39,41.41 31.41)),((41.41 31.41,41.96 30.39,41.66 31.11,41.41 31.41)),((41.66 31.11,41.96 30.39,41.85 30.77,41.66 31.11)),((40 28,40.39 28.04,39.61 28.04,40 28)),((39.61 28.04,40.39 28.04,41.41 28.59,39.61 28.04)),((39.61 28.04,41.41 28.59,38.59 28.59,39.61 28.04)),((39.61 28.04,38.59 28.59,38.89 28.34,39.61 28.04)),((39.61 28.04,38.89 28.34,39.23 28.15,39.61 28.04)),((41.41 28.59,40.39 28.04,41.11 28.34,41.41 28.59)),((41.11 28.34,40.39 28.04,40.77 28.15,41.11 28.34)),((41.41 28.59,41.66 28.89,41.96 29.61,41.41 28.59)),((41.41 28.59,41.96 29.61,41.96 30.39,41.41 28.59)),((41.96 30.39,41.96 29.61,42 30,41.96 30.39)),((41.96 29.61,41.66 28.89,41.85 29.23,41.96 29.61)),((38.59 28.59,38.04 30.39,38.04 29.61,38.59 28.59)),((38.59 28.59,38.04 29.61,38.34 28.89,38.59 28.59)),((38.34 28.89,38.04 29.61,38.15 29.23,38.34 28.89)),((38.04 29.61,38.04 30.39,38 30,38.04 29.61)))",
		},
		{
			inputWKT:      "POLYGON ((0 0, 0 200, 180 200, 180 0, 0 0), (20 180, 160 180, 160 20, 152.625 146.75, 20 180), (30 160, 150 30, 70 90, 30 160))",
			inputWKB:      "010300000003000000050000000000000000000000000000000000000000000000000000000000000000006940000000000080664000000000000069400000000000806640000000000000000000000000000000000000000000000000050000000000000000003440000000000080664000000000000064400000000000806640000000000000644000000000000034400000000000146340000000000058624000000000000034400000000000806640040000000000000000003e4000000000000064400000000000c062400000000000003e40000000000080514000000000008056400000000000003e400000000000006440",
			expectedEdges: "MULTILINESTRING ((0 200,180 200),(0 0,0 200),(0 0,180 0),(180 0,180 200),(152.625 146.75,180 0),(152.625 146.75,180 200),(152.625 146.75,160 180),(160 180,180 200),(0 200,160 180),(20 180,160 180),(0 200,20 180),(20 180,30 160),(0 200,30 160),(0 0,30 160),(30 160,70 90),(0 0,70 90),(70 90,150 30),(0 0,150 30),(150 30,160 20),(0 0,160 20),(160 20,180 0),(152.625 146.75,160 20),(150 30,152.625 146.75),(70 90,152.625 146.75),(30 160,152.625 146.75),(30 160,160 180))",
			expectedTris:  "MULTIPOLYGON (((0 200,0 0,30 160,0 200)),((0 200,30 160,20 180,0 200)),((0 200,20 180,160 180,0 200)),((0 200,160 180,180 200,0 200)),((180 200,160 180,152.625 146.75,180 200)),((180 200,152.625 146.75,180 0,180 200)),((0 0,180 0,160 20,0 0)),((0 0,160 20,150 30,0 0)),((0 0,150 30,70 90,0 0)),((0 0,70 90,30 160,0 0)),((30 160,70 90,152.625 146.75,30 160)),((30 160,152.625 146.75,160 180,30 160)),((30 160,160 180,20 180,30 160)),((152.625 146.75,70 90,150 30,152.625 146.75)),((152.625 146.75,150 30,160 20,152.625 146.75)),((152.625 146.75,160 20,180 0,152.625 146.75)))",
		},
	}

	for i, tc := range testcases {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tc))
	}
}
