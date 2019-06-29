package quadedge

type QuadEdge struct {
	initialized bool
	e           [4]Edge
}

func NewQEdge() *QuadEdge {
	var qe QuadEdge
	qe.e[0].num, qe.e[1].num, qe.e[2].num, qe.e[3].num = 0, 1, 2, 3
	qe.e[0].qe, qe.e[1].qe, qe.e[2].qe, qe.e[3].qe = &qe, &qe, &qe, &qe

	qe.e[0].next = &(qe.e[0])
	qe.e[1].next = &(qe.e[3])
	qe.e[2].next = &(qe.e[2])
	qe.e[3].next = &(qe.e[1])

	qe.initialized = true

	return &qe
}
