package subdivision

import (
	"context"
	"log"
	"testing"

	"github.com/go-spatial/geom/encoding/wkt"
	"github.com/go-spatial/geom/winding"

	"github.com/go-spatial/geom/planar/triangulate/delaunay/quadedge"
	"github.com/go-spatial/geom/planar/triangulate/delaunay/test/must"

	"github.com/go-spatial/geom"
)

func TestNewForPoints(t *testing.T) {
	type tcase struct {
		Name   string
		Desc   string
		Order  winding.Order
		Points [][2]float64
		Lines  []geom.Line
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {

			//if debug {
			log.Printf("Running test %v", t.Name())
			log.Printf("initial points\n%v", wkt.MustEncode(geom.MultiPoint(tc.Points)))
			//}
			sd, err := NewForPoints(context.Background(), tc.Order, tc.Points)
			if err != nil {
				t.Errorf("err, expected nil got %v", err)
				t.Logf("points: %v", wkt.MustEncode(geom.MultiPoint(tc.Points)))
				if err1, ok := err.(quadedge.ErrInvalid); ok {
					for i, estr := range err1 {
						t.Logf("%v: %v", i, estr)
					}
				}
			}

			err = sd.Validate(context.Background())
			if err != nil {
				t.Logf("points: %v", wkt.MustEncode(geom.MultiPoint(tc.Points)))
				if err1, ok := err.(quadedge.ErrInvalid); ok {
					for i, estr := range err1 {
						t.Logf("%03v : %v", i, estr)
					}
					dumpSD(t, sd)
				}
				t.Errorf(err.Error())
				return
			}
			var allLines []geom.Line
			err = sd.WalkAllEdges(func(e *quadedge.Edge) error {
				allLines = append(allLines, e.AsLine())
				return nil
			})
			if len(tc.Lines) != len(allLines) {
				dumpSD(t, sd)
				t.Errorf("lines, expected %v got %v", len(tc.Lines), len(allLines))
				return
			}
			//allLines = must.ParseMultilines([]byte(wkt.MustEncode(allLines)))
			seen := map[int]bool{}
			var didNotFind []int
		TESTCASE_LINES:
			for i, ln := range tc.Lines {
				for j, aln := range allLines {
					if seen[j] {
						continue
					}
					if cmp.PointEqual(ln[0], aln[0]) && cmp.PointEqual(ln[1], aln[1]) || // compare the start:start and end:end
						cmp.PointEqual(ln[0], aln[1]) && cmp.PointEqual(ln[1], aln[0]) { // compare the start:end and end:start
						seen[j] = true
						continue TESTCASE_LINES
					}
				}
				didNotFind = append(didNotFind, i)
			}
			if len(didNotFind) > 0 {
				t.Errorf("did not find lines, expected 0 got %v", len(didNotFind))
				for _, i := range didNotFind {
					t.Logf("Did not find: %v", wkt.MustEncode(tc.Lines[i]))
				}
				t.Logf("Got:\n%v", wkt.MustEncode(allLines))
			}
			for i := range allLines {
				if seen[i] {
					continue
				}
				t.Logf("Did not find: %v", wkt.MustEncode(allLines[i]))

			}
		}
	}

	tests := []tcase{
		{
			Desc: "one point",
			Points: [][2]float64{
				{0, 0},
			},
			Lines: must.ParseMultilines([]byte(
				`MULTILINESTRING ((-10 -10,0 0),(0 0,0 10),(0 10,-10 -10),(0 10,10 -10),(10 -10,-10 -10),(10 -10,0 0))`,
			)),
		},
		{
			Desc: "two points",
			Points: [][2]float64{
				{0, 0},
				{0, -6},
			},
			Lines: must.ParseMultilines([]byte(
				`MULTILINESTRING ((-10 -76,0 -6),(0 -6,0 0),(0 0,-10 -76),(0 0,0 130),(0 130,-10 -76),(0 130,10 -76),(10 -76,-10 -76),(10 -76,0 -6),(10 -76,0 0))`,
			)),
		},
		{
			Desc: "three points",
			Points: [][2]float64{
				{0, 0},
				{0, -6},
				{-6, 6},
			},
			Lines: must.ParseMultilines([]byte(
				`MULTILINESTRING ((-46 -136,-6 6),(-6 6,-3 256),(-3 256,-46 -136),(-3 256,40 -136),(40 -136,-46 -136),(40 -136,0 -6),(0 -6,-46 -136),(0 -6,-6 6),(0 -6,0 0),(0 0,-6 6),(0 0,-3 256),(0 0,40 -136))`,
			)),
		},
		{
			Desc: "four points",
			Points: [][2]float64{
				{0, 0},
				{0, -6},
				{-6, 6},
				{6, 6},
			},
			Lines: must.ParseMultilines([]byte(
				`MULTILINESTRING ((76 -136,6 6),(6 6,0 -6),(0 -6,76 -136),(0 -6,-76 -136),(-76 -136,76 -136),(-76 -136,0 256),(0 256,76 -136),(0 256,6 6),(0 256,-6 6),(-6 6,6 6),(-6 6,0 0),(0 0,6 6),(0 0,0 -6),(-6
 6,0 -6),(-6 6,-76 -136))`,
			)),
		},
		{
			Desc: "colinear folinear",
			Points: [][2]float64{
				{30, 4},
				{20, 4},
				{20, 2},
				{20, 6},
				{10, 6},
				{15, 4},
				{17, 4},
				{19, 3},
			},
			Lines: must.ReadMultilines("testdata/colinear_folinear.lines"),
		},
		{
			Desc:   "trunc something wrong with Florida",
			Points: must.ReadPoints("testdata/florida_trucated.points"),
			Lines: must.ParseMultilines([]byte(
				`MULTILINESTRING ((32 2576,33 2574),(33 2574,32 2583),(32 2583,32 2576),(32 2583,31 2581),(31 2581,32 2576),(31 2581,31 2578),(31 2578,32 2576),(31 2578,31 2577),(31 2577,32 2576),(31 2577,29 2571),(29 2571,32 2576),(29 2571,33 2574),(29 2571,33 2572),(33 2572,33 2574),(33 2572,38 2564),(38 2564,33 2574),(38 2564,61 2606),(61 2606,33 2574),(61 2606,32 2583),(61 2606,32 2584),(32 2584,32 2583),(32 2584,31 2582),(31 2582,32 2583),(31 2582,31 2581),(31 2582,30 2581),(30 2581,31 2581),(30 2581,30 2580),(30 2580,31 2581),(30 2580,30 2579),(30 2579,31 2581),(30 2579,31 2578),(30 2579,31 2577),(30 2579,26 2574),(26 2574,31 2577),(26 2574,29 2571),(26 2574,26 2572),(26 2572,29 2571),(26 2572,27 2571),(27 2571,29 2571),(27 2571,33 2564),(33 2564,29 2571),(33 2564,33 2572),(33 2564,36 2564),(36 2564,33 2572),(36 2564,37 2564),(37 2564,33 2572),(37 2564,38 2564),(37 2564,37 2562),(37 2562,38 2564),(37 2562,39 2562),(39 2562,38 2564),(39 2562,61 2606),(39 2562,63 2607),(63 2607,61 2606),(63 2607,61 2608),(61 2608,61 2606),(61 2608,49 2615),(49 2615,61 2606),(49 2615,47 2615),(47 2615,61 2606),(47 2615,32 2584),(47 2615,17 2612),(17 2612,32 2584),(17 2612,13 2608),(13 2608,32 2584),(13 2608,10 2600),(10 2600,32 2584),(10 2600,10 2596),(10 2596,32 2584),(10 2596,25 2579),(25 2579,32 2584),(25 2579,30 2581),(30 2581,32 2584),(25 2579,30 2580),(25 2579,30 2579),(25 2579,26 2574),(25 2579,22 2578),(22 2578,26 2574),(22 2578,24 2573),(24 2573,26 2574),(24 2573,26 2572),(24 2573,26 2571),(26 2571,26 2572),(26 2571,27 2571),(26 2571,23 2562),(23 2562,27 2571),(23 2562,33 2564),(23 2562,26 2559),(26 2559,33 2564),(26 2559,33 2559),(33 2559,33 2564),(33 2559,33 2560),(33 2560,33 2564),(33 2560,35 2563),(35 2563,33 2564),(35 2563,36 2564),(35 2563,36 2556),(36 2556,36 2564),(36 2556,36 2561),(36 2561,36 2564),(36 2561,37 2562),(37 2562,36 2564),(36 2561,38 2559),(38 2559,37 2562),(38 2559,39 2562),(38 2559,64 2446.875),(64 2446.875,39 2562),(64 2446.875,64 2607.500),(64 2607.500,39 2562),(64 2607.500,63 2607),(64 2607.500,61 2608),(64 2607.500,61 2610),(61 2610,61 2608),(61 2610,49 2615),(61 2610,54 2618),(54 2618,49 2615),(54 2618,49 2616),(49 2616,49 2615),(49 2616,47 2615),(49 2616,46 2617),(46 2617,47 2615),(46 2617,17 2612),(46 2617,18 2618),(18 2618,17 2612),(18 2618,11 2612),(11 2612,17 2612),(11 2612,13 2608),(11 2612,12 2607),(12 2607,13 2608),(12 2607,10 2600),(12 2607,8 2603),(8 2603,10 2600),(8 2603,8 2598),(8 2598,10 2600),(8 2598,9 2597),(9 2597,10 2600),(9 2597,10 2596),(9 2597,9 2590),(9 2590,10 2596),(9 2590,25 2579),(9 2590,22 2578),(9 2590,9 2586),(9 2586,22 2578),(9 2586,12 2575),(12 2575,22 2578),(12 2575,17 2574),(17 2574,22 2578),(17 2574,22 2573),(22 2573,22 2578),(22 2573,24 2573),(22 2573,26 2571),(22 2573,23 2562),(22 2573,16 2572),(16 2572,23 2562),(16 2572,21 2560),(21 2560,23 2562),(21 2560,24 2559),(24 2559,23 2562),(24 2559,26 2559),(24 2559,23 2558),(23 2558,26 2559),(23 2558,30 2553),(30 2553,26 2559),(30 2553,33 2559),(30 2553,36 2556),(36 2556,33 2559),(36 2556,34 2559),(34 2559,33 2559),(34 2559,33 2560),(34 2559,35 2563),(30 2553,34 2552),(34 2552,36 2556),(34 2552,38 2559),(38 2559,36 2556),(34 2552,64 2446.875),(34 2552,33 2550),(33 2550,64 2446.875),(33 2550,32 2471),(32 2471,64 2446.875),(32 2471,61 2448),(61 2448,64 2446.875),(61 2448,64 2407.692),(64 2407.692,64 2446.875),(64 2407.692,394 -15430),(394 -15430,64 2446.875),(394 -15430,64 2607.500),(394 -15430,64 4160),(64 4160,64 2607.500),(64 4160,44 2652),(44 2652,64 2607.500),(44 2652,49 2636),(49 2636,64 2607.500),(49 2636,53 2621),(53 2621,64 2607.500),(53 2621,61 2610),(53 2621,54 2618),(53 2621,50 2620),(50 2620,54 2618),(50 2620,49 2616),(50 2620,46 2617),(50 2620,46 2627),(46 2627,46 2617),(46 2627,46 2630),(46 2630,46 2617),(46 2630,45 2624),(45 2624,46 2617),(45 2624,45 2622),(45 2622,46 2617),(45 2622,45 2620),(45 2620,46 2617),(45 2620,18 2618),(45 2622,18 2618),(45 2624,18 2618),(45 2624,15 2624),(15 2624,18 2618),(15 2624,9 2623),(9 2623,18 2618),(9 2623,11 2612),(9 2623,6 2611),(6 2611,11 2612),(6 2611,9 2607),(9 2607,11 2612),(9 2607,12 2607),(9 2607,8 2603),(9 2607,6 2606),(6 2606,8 2603),(6 2606,6 2600),(6 2600,8 2603),(6 2600,7 2599),(7 2599,8 2603),(7 2599,8 2591),(8 2591,8 2603),(8 2591,8 2598),(8 2591,9 2597),(8 2591,9 2586),(9 2586,9 2597),(8 2591,8 2584),(8 2584,9 2586),(8 2584,12 2575),(8 2584,8 2578),(8 2578,12 2575),(8 2578,10 2574),(10 2574,12 2575),(10 2574,16 2572),(16 2572,12 2575),(16 2572,17 2574),(10 2574,14 2559),(14 2559,16 2572),(14 2559,21 2560),(14 2559,18 2558),(18 2558,21 2560),(18 2558,21 2557),(21 2557,21 2560),(21 2557,23 2558),(23 2558,21 2560),(21 2557,30 2553),(21 2557,33 2550),(33 2550,30 2553),(18 2558,33 2550),(14 2559,33 2550),(14 2559,32 2471),(14 2559,2 2453),(2 2453,32 2471),(2 2453,6 2454),(6 2454,32 2471),(6 2454,30 2467),(30 2467,32 2471),(30 2467,42 2454),(42 2454,32 2471),(42 2454,61 2448),(42 2454,48 2449),(48 2449,61 2448),(48 2449,64 2407.692),(48 2449,31 2439),(31 2439,64 2407.692),(31 2439,31 2413),(31 2413,64 2407.692),(31 2413,60 2404),(60 2404,64 2407.692),(60 2404,55 2398),(55 2398,64 2407.692),(55 2398,48 2388),(48 2388,64 2407.692),(48 2388,44 2383),(44 2383,64 2407.692),(44 2383,394 -15430),(44 2383,40 2380),(40 2380,394 -15430),(40 2380,34 2380),(34 2380,394 -15430),(34 2380,-330 -15430),(-330 -15430,394 -15430),(-330 -15430,32 39770),(32 39770,394 -15430),(32 39770,64 4160),(32 39770,3 2609),(3 2609,64 4160),(3 2609,10 2627),(10 2627,64 4160),(10 2627,14 2635),(14 2635,64 4160),(14 2635,15 2636),(15 2636,64 4160),(15 2636,20 2641),(20 2641,64 4160),(20 2641,23 2644),(23 2644,64 4160),(23 2644,36 2651),(36 2651,64 4160),(36 2651,41 2653),(41 2653,64 4160),(41 2653,44 2652),(41 2653,43 2651),(43 2651,44 2652),(43 2651,43 2646),(43 2646,44 2652),(43 2646,48 2638),(48 2638,44 2652),(48 2638,49 2636),(48 2638,48 2634),(48 2634,49 2636),(48 2634,52 2622),(52 2622,49 2636),(52 2622,53 2621),(52 2622,50 2620),(52 2622,46 2627),(52 2622,47 2632),(47 2632,46 2627),(47 2632,46 2630),(47 2632,44 2642),(44 2642,46 2630),(44 2642,26 2645),(26 2645,46 2630),(26 2645,18 2638),(18 2638,46 2630),(18 2638,45 2624),(18 2638,15 2624),(18 2638,15 2636),(15 2636,15 2624),(18 2638,20 2641),(26 2645,20 2641),(26 2645,23 2644),(26 2645,36 2651),(26 2645,43 2644),(43 2644,36 2651),(43 2644,43 2646),(43 2646,36 2651),(43 2646,42 2649),(42 2649,36 2651),(42 2649,41 2653),(42 2649,43 2651),(43 2644,44 2642),(44 2642,43 2646),(44 2642,48 2638),(44 2642,48 2634),(47 2632,48 2634),(14 2635,15 2624),(10 2627,15 2624),(10 2627,9 2623),(3 2609,9 2623),(3 2609,6 2611),(3 2609,3 2608),(3 2608,6 2611),(3 2608,6 2600),(6 2600,6 2611),(3 2608,6 2593),(6 2593,6 2600),(6 2593,6 2595),(6 2595,6 2600),(6 2595,7 2599),(6 2595,8 2591),(6 2595,7 2592),(7 2592,8 2591),(7 2592,4 2582),(4 2582,8 2591),(4 2582,7 2583),(7 2583,8 2591),(7 2583,8 2578),(8 2578,8 2591),(4 2582,8 2578),(4 2582,5 2572),(5 2572,8 2578),(5 2572,10 2574),(5 2572,14 2559),(5 2572,0 2451),(0 2451,14 2559),(0 2451,2 2453),(0 2451,6 2454),(0 2451,13 2442),(13 2442,6 2454),(13 2442,19 2455),(19 2455,6 2454),(19 2455,30 2467),(19 2455,29 2463),(29 2463,30 2467),(29 2463,42 2454),(29 2463,29 2460),(29 2460,42 2454),(29 2460,30 2457),(30 2457,42 2454),(30 2457,32 2453),(32 2453,42 2454),(32 2453,31 2439),(31 2439,42 2454),(32 2453,19 2455),(19 2455,31 2439),(19 2455,28 2437),(28 2437,31 2439),(28 2437,27 2435),(27 2435,31 2439),(27 2435,27 2431),(27 2431,31 2439),(27 2431,31 2413),(27 2431,25 2430),(25 2430,31 2413),(25 2430,21 2429),(21 2429,31 2413),(21 2429,25 2408),(25 2408,31 2413),(25 2408,48 2388),(48 2388,31 2413),(25 2408,34 2380),(34 2380,48 2388),(34 2380,44 2383),(25 2408,26 2381),(26 2381,34 2380),(26 2381,-330 -15430),(26 2381,17 2383),(17 2383,-330 -15430),(17 2383,10 2386),(10 2386,-330 -15430),(10 2386,0 2394),(0 2394,-330 -15430),(0 2394,0 2396),(0 2396,-330 -15430),(0 2396,0 2451),(0 2451,-330 -15430),(0 2451,32 39770),(0 2451,3 2609),(0 2451,3 2608),(0 2451,4 2582),(4 2582,3 2608),(4 2582,6 2593),(0 2396,3 2404),(3 2404,0 2451),(3 2404,5 2406),(5 2406,0 2451),(5 2406,13 2442),(5 2406,18 2436),(18 2436,13 2442),(18 2436,28 2437),(28 2437,13 2442),(18 2436,27 2435),(18 2436,25 2430),(25 2430,27 2435),(18 2436,21 2429),(5 2406,21 2429),(5 2406,13 2408),(13 2408,21 2429),(13 2408,15 2408),(15 2408,21 2429),(15 2408,19 2407),(19 2407,21 2429),(19 2407,25 2408),(19 2407,21 2406),(21 2406,25 2408),(21 2406,26 2381),(21 2406,17 2383),(21 2406,10 2386),(19 2407,10 2386),(15 2408,10 2386),(13 2408,10 2386),(13 2408,2 2393),(2 2393,10 2386),(2 2393,0 2394),(2 2393,0 2396),(2 2393,3 2404),(2 2393,5 2406),(32 2453,26 2457),(26 2457,19 2455),(26 2457,29 2463),(26 2457,29 2460),(26 2457,30 2457),(26 2457,30 2456),(30 2456,30 2457),(30 2456,32 2453),(7 2592,6 2593),(55 2398,31 2413),(6 2606,6 2611))
`,
			)),
		},
		{
			Desc:   "intersecting_lines_circle_inclusion_rounding_issue",
			Points: must.ReadPoints("testdata/florida_trucated_2.points"),
			Lines: must.ParseMultilines([]byte(
				`MULTILINESTRING ((-26 -2939,-1 30),(-1 30,0.500 5420),(0.500 5420,-26 -2939),(0.500 5420,27 -2939),(27 -2939,-26 -2939),(27 -2939,-1 -239),(-1 -239,-26 -2939),(-1 -239,-1 30),(-1 -239,0 -2),(0 -2,-1 30),(0 -2,0 0),(0 0,-1 30),(0 0,0 2),(0 2,-1 30),(0 2,2 -7),(2 -7,-1 30),(2 -7,0.500 5420),(2 -7,27 -2939),(2 -7,-1 -239),(2 -7,0 -2),(2 -7,0 0))
`,
			)),
		},
		{
			Desc: "bad_external_point",
			Points: must.PrasePoints([]byte(
				`MULTIPOINT ((19361630.4180414 6935925.51090632),(19361632.3327366 6935972.56790791),(19361538.4792739 6935976.39279552),(19361536.5645787 6935929.33577147))`,
			)),
		},
		{
			Desc: "bad_external_point_full",
			Points: must.PrasePoints([]byte(
				`MULTIPOINT((19359932.3028604 6936823.42893856),(19360250.8689878 6936688.69960884) ,(19360304.4693226 6936851.40220771) ,(19360304.5746052 6936851.40848484) ,(19361268.2934265 6936851.40848484) ,(19361267.6612167 6936835.82513453) ,(19361430.1208816 6936782.05044165) ,(19361432.9421194 6936851.40848484) ,(19361478.7453836 6936851.40848484) ,(19361709.8667619 6936723.53133297) ,(19361704.1226762 6936582.32953635) ,(19361798.0206667 6936578.52274976) ,(19361833.4647926 6936294.26940364) ,(19361739.589066 6936298.09444482) ,(19361686.9349468 6936158.86780031) ,(19361780.8106734 6936155.00604701) ,(19361775.0665877 6936013.86943516) ,(19361960.9033456 6935959.12564909) ,(19361964.7104722 6936053.2403234) ,(19362226.6118382 6936136.80066362) ,(19362162.0465336 6936280.82660515) ,(19362023.1420729 6936333.6600593) ,(19362223.3279132 6936631.76275541) ,(19362454.7247934 6936752.03059318) ,(19362454.7247934 6934328.98655178) ,(19359932.3028604 6934328.98655178) ,(19359932.3028604 6936823.42893856) ,(19361352.631384 6936031.09992153) ,(19361414.8478475 6936405.63769592) ,(19361274.9749073 6936434.78569641) ,(19361178.2382698 6936368.1224444) ,(19361244.7405336 6936271.15368887) ,(19361305.6990867 6936033.01237908) ,(19361352.631384 6936031.09992153) ,(19361630.4180414 6935925.51090632) ,(19361632.3327366 6935972.56790791) ,(19361538.4792739 6935976.39279552) ,(19361536.5645787 6935929.33577147) ,(19361630.4180414 6935925.51090632) ,(19362086.3826757 6935576.95934416) ,(19362049.0127226 6935814.16767519) ,(19362002.1026892 6935816.08008097) ,(19361998.2732987 6935721.968192) ,(19361951.3298694 6935723.86218747) ,(19361945.5857837 6935582.733169) ,(19362086.3826757 6935576.95934416))
`,
			)),
		},
		{
			Desc:   "something wrong with Florida",
			Points: must.ReadPoints("testdata/florida.points"),
			Lines:  must.ReadMultilines("testdata/florida_expected.lines"),
		},
		{
			Desc:   "something wrong with north Africa",
			Points: must.ReadPoints("testdata/north_africa.points"),
		},
		{
			Desc:   "intersecting lines are generated 1",
			Points: must.ReadPoints("testdata/intersecting_lines_1.points"),
			Lines:  must.ReadMultilines("testdata/intersecting_lines_1_expected.lines"),
		},
		{
			Desc:   "counter clockwise error east of china",
			Points: must.ReadPoints("testdata/east_of_china.points"),
		},
		{
			Desc:   "error failed to insert point 8",
			Points: [][2]float64{[2]float64{-1.3625395451e+07, 4.551405984e+06}, [2]float64{-1.3625385953e+07, 4.551392498e+06}, [2]float64{-1.3625144745e+07, 4.551583426e+06}, [2]float64{-1.3625317363e+07, 4.55141451e+06}, [2]float64{-1.3625204228e+07, 4.551495519e+06}, [2]float64{-1.3625225288e+07, 4.551499794e+06}, [2]float64{-1.3625218504e+07, 4.55149004e+06}, [2]float64{-1.3625167969e+07, 4.551553549e+06}, [2]float64{-1.3625206458e+07, 4.551498625e+06}, [2]float64{-1.3625137934e+07, 4.551573731e+06}},
			//Points: must.ReadPoints("testdata/failed_to_insert_point_8_points.wkt"),
			Lines: must.ReadMultilines("testdata/failed_to_insert_point_8_lines.wkt"),
		},
		{
			Desc:   "issue 96 1",
			Points: must.ReadPoints("testdata/issue/96/points_1.wkt"),
			Lines:  must.ReadMultilines("testdata/issue/96/lines_1.wkt"),
		},
		{
			Desc:   "issue 96 2",
			Points: must.ReadPoints("testdata/issue/96/points_2.wkt"),
			//			Lines:  must.ReadMultilines("testdata/issue/96/lines_1.wkt"),
		},
		{
			Desc:   "issue 96 simplified",
			Points: must.ReadPoints("testdata/issue/96/points_simplified.wkt"),
			//			Lines:  must.ReadMultilines("testdata/issue/96/lines_1.wkt"),
		},
	}

	//t.Skip("TestForNewPoints disabled")
	//return

	for _, tc := range tests {
		t.Run(tc.Desc, fn(tc))
	}
}

func TestNewSubdivision(t *testing.T) {
	// This is not going to be a table driven test. It going to test
	// one thing, that is a geom.Triangle{0 0,10 0,5 10} to make sure
	// that a subdivision can be make from such a triangle.
	tri := geom.Triangle{{0, 0}, {10, 0}, {5, 10}}
	var order winding.Order

	sd := New(order, geom.Point(tri[0]), geom.Point(tri[1]), geom.Point(tri[2]))
	if sd.ptcount != 3 {
		t.Errorf("ptcount, expected 3, got %v", sd.ptcount)
	}

	if !cmp.GeomPointEqual(sd.frame[0], geom.Point(tri[0])) {
		t.Errorf("frame point 0, expected %v, got %v", geom.Point(tri[0]), sd.frame[0])
	}
	if !cmp.GeomPointEqual(sd.frame[1], geom.Point(tri[1])) {
		t.Errorf("frame point 0, expected %v, got %v", geom.Point(tri[1]), sd.frame[1])
	}
	if !cmp.GeomPointEqual(sd.frame[2], geom.Point(tri[2])) {
		t.Errorf("frame point 0, expected %v, got %v", geom.Point(tri[2]), sd.frame[2])
	}

	if sd.startingEdge == nil {
		t.Errorf("starting edge, expected non-nil, got nil")
	}
	seln := sd.startingEdge.AsLine()
	exln := geom.Line{{0, 0}, {10, 0}}
	if !cmp.LineEqual(seln, exln) {
		t.Errorf("starting edge, expected %v got %v ", exln, seln)
	}

	// Need to validate each edge
	// Let's see if we can find all the edges
	edges := make([]geom.Line, 0, 3)
	sd.WalkAllEdges(func(e *quadedge.Edge) error {
		edges = append(edges, e.AsLine())
		return nil
	})
	if len(edges) != 3 {
		t.Errorf("number of edges, expected 3 got %v", len(edges))
		for i := range edges {
			t.Logf("\tEdge[%v]: %v", i, wkt.MustEncode(edges[i]))
		}
	}
	expectedEdges := []geom.Line{
		{{0, 0}, {10, 0}},
		{{10, 0}, {5, 10}},
		{{5, 10}, {0, 0}},
	}

	for i := range edges {
		if !cmp.LineEqual(edges[i], expectedEdges[i]) {
			t.Errorf("edge %v, expected %v got %v", i, wkt.MustEncode(expectedEdges[i]), wkt.MustEncode(edges[i]))
		}
	}
	if err := sd.Validate(context.Background()); err != nil {
		t.Errorf("validate, expected nil , got %v", err)
		if err1, ok := err.(quadedge.ErrInvalid); ok {
			for i := range err1 {
				t.Logf("err[%v]: %v", i, err1[i])
			}
		}
	}
}

func genAndTestNewSD(t *testing.T, order winding.Order, trianglePoint [3]geom.Point) (*Subdivision, []*quadedge.Edge) {
	triangleEdge := make([]*quadedge.Edge, len(trianglePoint))
	sd := New(order, trianglePoint[0], trianglePoint[1], trianglePoint[2])
	se := sd.startingEdge
	if !cmp.GeomPointEqual(*(se.Orig()), trianglePoint[0]) {
		se = se.FindONextDest(trianglePoint[0]).Sym()
	}

	triangleEdge[0] = se.FindONextDest(trianglePoint[2])
	triangleEdge[1] = triangleEdge[0].ONext().Sym()
	triangleEdge[2] = triangleEdge[0].Sym()

	for i := range trianglePoint {
		// verify that point edge is origin is correct
		if !cmp.GeomPointEqual(*(triangleEdge[i].Orig()), trianglePoint[i]) {
			t.Errorf("new edge %v origin, expected %v got %v", i, wkt.MustEncode(trianglePoint[i]), wkt.MustEncode(*(triangleEdge[i].Orig())))
			return nil, nil
		}
		// Let's verify that all vertex have only two edges
		if triangleEdge[i].ONext() == triangleEdge[i] || triangleEdge[i].ONext().ONext() != triangleEdge[i] {
			t.Errorf("vertex %v, expected two edges, got not two", i)
			return nil, nil
		}
	}
	return sd, triangleEdge
}

func testEdgeONextOPrevDest(t *testing.T, i string, e *quadedge.Edge, onextDest geom.Point, oprevDest geom.Point) {
	printedEdge := false
	if !cmp.GeomPointEqual(*(e.ONext().Dest()), onextDest) {
		t.Logf("edge %v", wkt.MustEncode(e.AsLine()))
		printedEdge = true
		t.Errorf("edge %v onext, expected %v got %v", i, onextDest, *(e.ONext().Dest()))
	}
	if !cmp.GeomPointEqual(*(e.OPrev().Dest()), oprevDest) {
		if !printedEdge {
			t.Logf("edge %v", wkt.MustEncode(e.AsLine()))
		}
		t.Errorf("edge %v oprev, expected %v got %v", i, oprevDest, *(e.OPrev().Dest()))
	}
}

func TestSubdivisionInsertSiteOnePoint(t *testing.T) {
	var (
		order         winding.Order
		trianglePoint = [...]geom.Point{
			geom.Point{-100, -100},
			geom.Point{0, 100},
			geom.Point{100, -100},
		}
		insertPoint = geom.Point{0, 0}
	)

	sd, triangleEdge := genAndTestNewSD(t, order, trianglePoint)
	if sd == nil {
		return
	}

	sd.InsertSite(insertPoint)

	for i := range trianglePoint {
		// Let's verify that all vertex have only two edges
		if triangleEdge[i].ONext() == triangleEdge[i] ||
			triangleEdge[i].ONext().ONext() == triangleEdge[i] ||
			triangleEdge[i].ONext().ONext().ONext() != triangleEdge[i] {
			t.Errorf("vertex %v, expected three edges, got not three", i)
		}
	}

	testEdgeONextOPrevDest(t, "0", triangleEdge[0], insertPoint, trianglePoint[1])
	testEdgeONextOPrevDest(t, "2", triangleEdge[2], trianglePoint[1], insertPoint)
	testEdgeONextOPrevDest(t, "1", triangleEdge[1], insertPoint, trianglePoint[2])

}
