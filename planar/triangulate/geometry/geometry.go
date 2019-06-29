package geometry


// CCW returns if the points a,b,c, are in a counterclockwise order
func CCW(a, b, c Point) bool {
	return TriArea(a, b, c) > 0
}

func Extent(pts ...[2]float64) (ext [2][2]float64) {
	if len(pts) == 0 {
		return ext
	}

	ext[0] = pts[0]
	ext[1] = pts[0]

	if len(pts) == 1 {
		return ext
	}

	for _, pt := range pts[1:] {
		if pt[0] < ext[0][0] {
			ext[0][0] = pt[0]
		}
		if pt[0] > ext[1][0] {
			ext[1][0] = pt[0]
		}
		if pt[1] < ext[0][1] {
			ext[0][1] = pt[1]
		}
		if pt[1] > ext[1][1] {
			ext[1][1] = pt[1]
		}
	}
	return ext
}

func TriangleContaining(pts ...[2]float64) (tri [3][2]float64) {
	const buff = 10
	ext := Extent(pts...)
	xlen := ext[1][0] - ext[0][0]
	ylen := ext[1][1] - ext[0][1]
	x2len := xlen / 2

	nx := ext[0][0] - (x2len * buff)
	cx := ext[0][0] + x2len
	xx := ext[1][0] + (x2len * buff)

	ny := ext[0][1] - (ylen * buff)
	xy := ext[1][1] + (2 * ylen * buff)

	tri[0] = [2]float64{nx, ny}
	tri[1] = [2]float64{cx, xy}
	tri[2] = [2]float64{xx, ny}

	return tri
}

func AppendNonRepeat(pts []Point, v Point) []Point {
	if len(pts) == 0 || !ArePointsEqual(pts[len(pts)-1],v) {
		return append(pts, v)
	}
	return pts
}
