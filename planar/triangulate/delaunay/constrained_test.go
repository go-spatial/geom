package delaunay_test

import (
	"context"
	"math"
	"testing"

	"github.com/gdey/errors"
	"github.com/go-spatial/geom/encoding/wkt"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/planar/triangulate/delaunay"
	"github.com/go-spatial/geom/planar/triangulate/delaunay/quadedge"
	"github.com/go-spatial/geom/planar/triangulate/delaunay/subdivision"
	"github.com/go-spatial/geom/planar/triangulate/delaunay/test/must"
)

func TestConstraint(t *testing.T) {

	fn := func(tc delaunay.Constrained) func(*testing.T) {
		return func(t *testing.T) {
			if !delaunay.EnableConstraints {
				t.Skipf("constraints not enabled.")
				return
			}

			var err error

			pts := tc.Points
			for _, ct := range tc.Constraints {
				pts = append(pts, ct[0], ct[1])
			}
			var sd *subdivision.Subdivision

			ctx := context.Background()
			sd, err = subdivision.NewForPoints(ctx, tc.Order, pts)
			if err != nil {
				t.Errorf("error, expected nil, got %v", err)
				return
			}
			isValid := sd.IsValid(ctx)

			if !isValid {
				t.Errorf("triangulation is not valid")
				dumpSD(t, sd)
				return
			}

			// Run the constraints as sub tests
			vxidx := sd.VertexIndex()
			total := len(tc.Constraints)

			for i, ct := range tc.Constraints {
				start, end := geom.Point(ct[0]), geom.Point(ct[1])
				if _, _, exists, _ := subdivision.ResolveStartingEndingEdges(sd.Order, vxidx, start, end); exists {
					// Nothing to do, edge already in the subdivision.
					continue
				}

				t.Logf("Adding Constraint %v of %v", i, total)

				bLine, bEdges := dumpSDWithinStr(sd, start, end)

				err := sd.InsertConstraint(ctx, vxidx, start, end)
				if err != nil {

					t.Logf("errored: %v", err)
					t.Logf("Before:\nLine:%v\nEdges:\n%v", bLine, bEdges)
					t.Logf("failed to add constraint %v of %v", i, total)
					dumpSDWithin(t, sd, start, end)
					t.Errorf("got err: %v", err)
					return
				}
				if err := sd.Validate(ctx); err != nil {
					t.Logf("Before:\nLine:%v\nEdges:\n%v", bLine, bEdges)
					t.Logf("Subdivision not valid")
					if er, ok := err.(quadedge.ErrInvalid); ok {
						for i, estr := range er {
							t.Logf("%03v: %v", i, estr)
						}
					}
					dumpSDWithin(t, sd, start, end)
					t.Errorf("Left subdivision in an invalid state")
					return
				}
			}

			dumpSD(t, sd)
			// Get all the triangles and check to see if they are what we expect
			// them to be.
			tri, err := sd.Triangles(false)
			if err != nil {
				t.Errorf("error, expected nil, got %v", err)
				return
			}
			t.Logf("Number of triangles: %v", len(tri))

		}
	}

	tests := packageTests

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}

func TestGeomConstrained(t *testing.T) {
	type tcase struct {
		Name         string
		Desc         string
		tri          delaunay.GeomConstrained
		includeFrame bool
		expTriangles []geom.Triangle
		err          error
		skip         string
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {

			if !delaunay.EnableConstraints {
				t.Skipf("constraints not enabled.")
				return
			}
			if tc.skip != "" {
				t.Skipf(tc.skip)
			}
			got, err := tc.tri.Triangles(context.Background(), tc.includeFrame)
			if tc.err != nil {
				if err == nil {
					t.Errorf("error, expected %v, got nil", tc.err)
					return
				}
				t.Logf("got error:%v", err)
			}
			if len(got) != len(tc.expTriangles) {
				t.Errorf("number of triangles, expected %v got %v", len(tc.expTriangles), len(got))
				t.Logf("Got the following triangles:")
				for i := range got {
					t.Logf("%03v %v", i, wkt.MustEncode(got[i]))
				}
				return
			}
		}
	}
	tests := []tcase{
		{
			Desc: "empty",
		},
		{
			Desc: "empty",
			tri: delaunay.GeomConstrained{
				Points: []geom.Point{{0, 0}, {0, 0}, {0, 0}, {0, 0}},
			},
		},
		{
			Desc: "empty",
			tri: delaunay.GeomConstrained{
				Constraints: []geom.Line{{{0, 0}, {0, 0}}},
			},
		},
		{
			Desc: "empty",
			tri:  delaunay.GeomConstrained{},
		},
		{
			Desc: "bad lines",
			tri: delaunay.GeomConstrained{
				Constraints: []geom.Line{
					{{math.Copysign(0, -1), 4096}, {math.Copysign(0, -1), 4096}},
					{{math.Copysign(0, -1), 4096}, {0, 4096}},
					{{math.Copysign(0, -1), 4096}, {0, 4096}},
					{{0, 4096}, {0, 4096}},
				},
			},
			err:  errors.String("invalid points/constraints"),
			skip: "bad lines not working",
		},

		{
			/*
				Github issue #70: https://github.com/go-spatial/geom/issues/70
				The main issue is one of the cutout (the one that looks like a rectangle)
				has an extra line sticking out of it. This extra point is encoded in the
				tile and not just a render issue. So, some how the make valid algo is adding
				the extra point
			*/
			Desc: "issue#70",
			tri: delaunay.GeomConstrained{
				Constraints: must.DecodeAsLines([]byte(
					`POLYGON ((19359932.3028604 6936823.42893856,19360250.8689878 6936688.69960884,19360304.4693226 6936851.40220771,19360304.5746052 6936851.40848484,19361268.2934265 6936851.40848484,19361267.6612167 6936835.82513453,19361430.1208816 6936782.05044165,19361432.9421194 6936851.40848484,19361478.7453836 6936851.40848484,19361709.8667619 6936723.53133297,19361704.1226762 6936582.32953635,19361798.0206667 6936578.52274976,19361833.4647926 6936294.26940364,19361739.589066 6936298.09444482,19361686.9349468 6936158.86780031,19361780.8106734 6936155.00604701,19361775.0665877 6936013.86943516,19361960.9033456 6935959.12564909,19361964.7104722 6936053.2403234,19362226.6118382 6936136.80066362,19362162.0465336 6936280.82660515,19362023.1420729 6936333.6600593,19362223.3279132 6936631.76275541,19362454.7247934 6936752.03059318,19362454.7247934 6934328.98655178,19359932.3028604 6934328.98655178,19359932.3028604 6936823.42893856),(19361352.631384 6936031.09992153,19361414.8478475 6936405.63769592,19361274.9749073 6936434.78569641,19361178.2382698 6936368.1224444,19361244.7405336 6936271.15368887,19361305.6990867 6936033.01237908,19361352.631384 6936031.09992153),(19361630.4180414 6935925.51090632,19361632.3327366 6935972.56790791,19361538.4792739 6935976.39279552,19361536.5645787 6935929.33577147,19361630.4180414 6935925.51090632),(19362086.3826757 6935576.95934416,19362049.0127226 6935814.16767519,19362002.1026892 6935816.08008097,19361998.2732987 6935721.968192,19361951.3298694 6935723.86218747,19361945.5857837 6935582.733169,19362086.3826757 6935576.95934416))`,
				)),
			},
			// TODO(gdey): Actually test the triangles are the triangles we expect. Instead of just the number of triangles
			expTriangles: make([]geom.Triangle, 73),
		},
		{
			Desc: "issue#70_full",
			tri: delaunay.GeomConstrained{
				Constraints: must.DecodeAsLines([]byte(
					`MULTILINESTRING ((19362352.506 6936749.190,19362352.506 6941769.160),(19362352.506 6936749.190,19362449.259 6936749.190),(19362352.506 6941769.160,19362352.506 6937378.338),(19362352.506 6941769.160,19367372.476 6941769.160),(19362352.506 6937378.338,19362352.506 6939242.939),(19362352.506 6937378.338,19362439.844 6937329.948),(19362352.506 6939242.939,19362352.506 6939148.626),(19362352.506 6939242.939,19362376.626 6939241.950),(19362352.506 6939148.626,19362372.774 6939147.798),(19362372.774 6939147.798,19362376.626 6939241.950),(19362439.844 6937329.948,19362583.547 6937394.835),(19362449.259 6936749.190,19362559.580 6936806.529),(19362449.259 6936749.190,19363552.512 6936749.190),(19362451.355 6938767.392,19362455.184 6938861.540),(19362451.355 6938767.392,19362543.338 6938716.493),(19362455.184 6938861.540,19362502.139 6938859.627),(19362502.139 6938859.627,19362543.338 6938716.493),(19362533.765 6938481.167,19362535.679 6938528.239),(19362533.765 6938481.167,19362721.561 6938473.478),(19362535.679 6938528.239,19362582.623 6938526.289),(19362559.580 6936806.529,19363186.130 6936933.205),(19362582.623 6938526.289,19362586.474 6938620.434),(19362583.547 6937394.835,19362622.820 6937204.699),(19362586.474 6938620.434,19362721.561 6938473.478),(19362622.820 6937204.699,19362900.651 6937099.021),(19362896.800 6937004.894,19362900.651 6937099.021),(19362896.800 6937004.894,19363186.130 6936933.205),(19363270.421 6936942.474,19363373.915 6937173.893),(19363270.421 6936942.474,19363552.512 6936749.190),(19363373.915 6937173.893,19363608.621 6937164.274),(19363552.512 6936749.190,19365051.369 6936749.190),(19363608.621 6937164.274,19363661.353 6937303.518),(19363661.353 6937303.518,19364081.907 6937239.073),(19364081.907 6937239.073,19364087.696 6937380.268),(19364087.696 6937380.268,19364134.628 6937378.319),(19364130.776 6937284.225,19364134.628 6937378.319),(19364130.776 6937284.225,19364224.663 6937280.345),(19364153.908 6937848.931,19364243.933 6937750.933),(19364153.908 6937848.931,19364296.653 6937890.206),(19364168.079 6937046.992,19364224.663 6937280.345),(19364168.079 6937046.992,19364639.428 6937074.726),(19364243.933 6937750.933,19364296.653 6937890.206),(19364639.428 6937074.726,19364647.154 6937262.964),(19364647.154 6937262.964,19364836.865 6937302.268),(19364768.670 6936786.575,19364836.865 6937302.268),(19364768.670 6936786.575,19365049.362 6936751.375),(19365049.362 6936751.375,19365051.369 6936749.190),(19365051.369 6936749.190,19365566.484 6936749.190),(19365452.918 6937418.266,19365456.792 6937512.398),(19365452.918 6937418.266,19365546.816 6937414.404),(19365456.792 6937512.398,19365533.268 6937478.133),(19365533.268 6937478.133,19365609.244 6937788.933),(19365546.816 6937414.404,19365642.629 6937457.589),(19365566.484 6936749.190,19365572.464 6936894.823),(19365566.484 6936749.190,19365919.039 6936749.190),(19365572.464 6936894.823,19365904.931 6936975.358),(19365609.244 6937788.933,19365656.176 6937786.983),(19365642.629 6937457.589,19365656.176 6937786.983),(19365904.931 6936975.358,19365908.793 6937069.448),(19365908.793 6937069.448,19366049.613 6937063.636),(19365919.039 6936749.190,19365920.560 6936786.152),(19365919.039 6936749.190,19367372.476 6936749.190),(19365920.560 6936786.152,19366086.871 6936826.409),(19366049.613 6937063.636,19366086.871 6936826.409),(19367372.476 6936749.190,19367372.476 6941769.160)), `,
				)),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Desc, fn(tc))
	}
}
