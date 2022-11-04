package goosm

import "math"

// Point list for find multiples points into db.
type PointList struct {
	// id do open street maps
	Id   int64  `bson:"id"`
	List []Node `bson:"list"`

	HasKeyValue bool `bson:"hasKeyValue" json:"-"`
}

func (el *PointList) GetId() int64 {
	return el.Id
}

func (el *PointList) AddFromPointList(point *PointList) {
	if len(el.List) == 0 {
		el.List = make([]Node, 0)
	}

	for _, p := range point.List {
		el.List = append(el.List, p)
	}
}

//func (el *PointList) AddPointLatLngDegrees(latitudeAFlt, longitudeAFlt float64) error {
//	if len(el.List) == 0 {
//		el.List = make([]Node, 0)
//	}
//
//	var p = Node{}
//	var err = p.SetLatLngDegrees(latitudeAFlt, longitudeAFlt)
//	if err != nil {
//		return err
//	}
//
//	el.List = append(el.List, p)
//
//	return nil
//}

//func (el *PointList) AddPointLatLngDecimalDrees(latitudeDegreesAFlt, latitudePrimesAFlt, latitudeSecondsAFlt, longitudeDegreesAFlt, longitudePrimesAFlt, longitudeSecondsAFlt int64) {
//	if len(el.List) == 0 {
//		el.List = make([]Node, 0)
//	}
//
//	var p = Node{}
//	p.SetLatLngDecimalDrees(latitudeDegreesAFlt, latitudePrimesAFlt, latitudeSecondsAFlt, longitudeDegreesAFlt, longitudePrimesAFlt, longitudeSecondsAFlt)
//	el.List = append(el.List, p)
//}

//func (el *PointList) AddPointLngLatDecimalDrees(longitudeDegreesAFlt, longitudePrimesAFlt, longitudeSecondsAFlt, latitudeDegreesAFlt, latitudePrimesAFlt, latitudeSecondsAFlt int64) {
//	if len(el.List) == 0 {
//		el.List = make([]Node, 0)
//	}
//
//	var p = Node{}
//	p.SetLngLatDecimalDrees(longitudeDegreesAFlt, longitudePrimesAFlt, longitudeSecondsAFlt, latitudeDegreesAFlt, latitudePrimesAFlt, latitudeSecondsAFlt)
//	el.List = append(el.List, p)
//}

func (el *PointList) AddPointLngLatDegrees(longitudeAFlt, latitudeAFlt float64) error {
	if len(el.List) == 0 {
		el.List = make([]Node, 0)
	}

	var p = Node{}
	var err = p.SetLngLatDegrees(longitudeAFlt, latitudeAFlt)
	if err != nil {
		return err
	}

	el.List = append(el.List, p)

	return nil
}

func (el *PointList) AddPointXYDegrees(xAFlt, yAFlt float64) error {
	if len(el.List) == 0 {
		el.List = make([]Node, 0)
	}

	var p = Node{}
	var err = p.SetLngLatDegrees(xAFlt, yAFlt)
	if err != nil {
		return err
	}

	el.List = append(el.List, p)

	return nil
}

//func (el *PointList) AddPointLatLngRadians(latitudeAFlt, longitudeAFlt float64) error {
//	if len(el.List) == 0 {
//		el.List = make([]Node, 0)
//	}
//
//	var p = Node{}
//	var err = p.SetLatLngRadians(latitudeAFlt, longitudeAFlt)
//	if err != nil {
//		return err
//	}
//
//	el.List = append(el.List, p)
//
//	return nil
//}

func (el *PointList) AddPointLngLatRadians(longitudeAFlt, latitudeAFlt float64) error {
	if len(el.List) == 0 {
		el.List = make([]Node, 0)
	}

	var p = Node{}
	var err = p.SetLngLatRadians(longitudeAFlt, latitudeAFlt)
	if err != nil {
		return err
	}

	el.List = append(el.List, p)

	return nil
}

func (el *PointList) AddPointXYRadians(xAFlt, yAFlt float64) error {
	if len(el.List) == 0 {
		el.List = make([]Node, 0)
	}

	var p = Node{}
	var err = p.SetLngLatRadians(xAFlt, yAFlt)
	if err != nil {
		return err
	}

	el.List = append(el.List, p)

	return nil
}

// Get all points after Find()
func (el *PointList) GetAll() []Node {
	return el.List
}

func (el *PointList) GetBox() (box Box, err error) {
	return GetBox(&el.List)
}

func (el *PointList) GetReverse() *PointList {
	for left, right := 0, len(el.List)-1; left < right; left, right = left+1, right-1 {
		el.List[left], el.List[right] = el.List[right], el.List[left]
	}

	return el
}

func (el *PointList) AddPoint(PointAStt Node) {
	el.List = append(el.List, PointAStt)
}

func (el *PointList) hullIsPointInsidePolygon(pointBStt Node, polygonAAStt []Node) bool {
	var i int
	var result = false
	var j = len(polygonAAStt) - 1

	for i = 0; i < len(polygonAAStt); i += 1 {
		if (polygonAAStt[i].Loc[1] < pointBStt.Loc[1] && polygonAAStt[j].Loc[1] > pointBStt.Loc[1]) || (polygonAAStt[j].Loc[1] < pointBStt.Loc[1] && polygonAAStt[i].Loc[1] > pointBStt.Loc[1]) {
			if polygonAAStt[i].Loc[0]+(pointBStt.Loc[1]-polygonAAStt[i].Loc[1])/(polygonAAStt[j].Loc[1]-polygonAAStt[i].Loc[1])*(polygonAAStt[j].Loc[0]-polygonAAStt[i].Loc[0]) < pointBStt.Loc[0] {
				result = !result
			}
		}
		j = i
	}

	return result
}

func (el *PointList) hullCheckEdgeIntersection(p0, p1, p2, p3 Node) bool {
	var s1_x = p1.Loc[0] - p0.Loc[0]
	var s1_y = p1.Loc[1] - p0.Loc[1]
	var s2_x = p3.Loc[0] - p2.Loc[0]
	var s2_y = p3.Loc[1] - p2.Loc[1]
	var s = (-s1_y*(p0.Loc[0]-p2.Loc[0]) + s1_x*(p0.Loc[1]-p2.Loc[1])) / (-s2_x*s1_y + s1_x*s2_y)
	var t = (s2_x*(p0.Loc[1]-p2.Loc[1]) - s2_y*(p0.Loc[0]-p2.Loc[0])) / (-s2_x*s1_y + s1_x*s2_y)

	return s > 0 && s < 1 && t > 0 && t < 1
}

func (el *PointList) hullCheckEdgeIntersectionList(hull []Node, curEdgeStart, curEdgeEnd, checkEdgeStart, checkEdgeEnd Node) bool {
	var i int
	var e1, e2 int
	var p1, p2 Node
	for i = 0; i < len(hull)-2; i += 1 {
		e1 = i
		e2 = i + 1
		p1 = hull[e1]
		p2 = hull[e2]

		if el.equality(curEdgeStart, p1) && el.equality(curEdgeEnd, p2) {
			continue
		}

		if el.hullCheckEdgeIntersection(checkEdgeStart, checkEdgeEnd, p1, p2) {
			return true
		}
	}
	return false
}

func (el *PointList) ConvertToConvexHull() {
	var hull = el.ConvexHull()
	el.List = hull.List
}

func (el *PointList) ConvexHull() PointList {
	var i, j, bot int
	var tmp = Node{}
	var P = make([]Node, len(el.List)) // = el.List
	var hull = PointList{}
	hull.List = make([]Node, 0)
	var minmin, minmax int
	var maxmin, maxmax int
	var xmin, xmax float64

	for k, v := range el.List {
		el.copyFrom(&P[k], &v)
	}

	// Sort P by x and y
	for i = 0; i < len(P); i += 1 {
		for j = i + 1; j < len(P); j += 1 {
			if P[j].Loc[0] < P[i].Loc[0] || (P[j].Loc[0] == P[i].Loc[0] && P[j].Loc[1] < P[i].Loc[1]) {
				el.copyFrom(&tmp, &P[i])
				el.copyFrom(&P[i], &P[j])
				el.copyFrom(&P[j], &tmp)
			}
		}
	}

	// the output array H[] will be used as the stack
	// i array scan index

	// Get the indices of points with min x-coord and min|max y-coord
	minmin = 0
	xmin = P[0].Loc[0]
	for i = 1; i < len(P); i += 1 {
		if P[i].Loc[0] != xmin {
			break
		}
	}

	minmax = i - 1
	if minmax == len(P)-1 { // degenerate case: all x-coords == xmin
		hull.List = append(hull.List, P[minmin])
		if P[minmax].Loc[1] != P[minmin].Loc[1] {
			hull.List = append(hull.List, P[minmax]) // a  nontrivial segment
		}
		hull.List = append(hull.List, P[minmin]) // add polygon endpoint
		return hull
	}

	// Get the indices of points with max x-coord and min|max y-coord
	maxmax = len(P) - 1
	xmax = P[len(P)-1].Loc[0]
	for i = len(P) - 2; i >= 0; i -= 1 {
		if P[i].Loc[0] != xmax {
			break
		}
	}
	maxmin = i + 1

	// Compute the lower hull on the stack H
	hull.List = append(hull.List, P[minmin]) // push  minmin point onto stack
	i = minmax
	for i+1 <= maxmin {
		i += 1

		// the lower line joins P[minmin]  with P[maxmin]
		if el.hullCcw(P[minmin], P[maxmin], P[i]) >= 0 && i < maxmin {
			continue // ignore P[i] above or on the lower line
		}

		for len(hull.List) > 1 { // there are at least 2 points on the stack
			// test if  P[i] is left of the line at the stack top
			if el.hullCcw(hull.List[len(hull.List)-2], hull.List[len(hull.List)-1], P[i]) > 0 {
				break // P[i] is a new hull  vertex
			}
			hull.List = hull.List[:len(hull.List)-1] // pop top point off  stack
		}
		hull.List = append(hull.List, P[i]) // push P[i] onto stack
	}

	// Next, compute the upper hull on the stack H above  the bottom hull
	if maxmax != maxmin { // if  distinct xmax points
		hull.List = append(hull.List, P[maxmax]) // push maxmax point onto stack
	}
	bot = len(hull.List) // the bottom point of the upper hull stack
	i = maxmin
	for (i - 1) >= minmax {
		i -= 1
		// the upper line joins P[maxmax]  with P[minmax]
		if el.hullCcw(P[maxmax], P[minmax], P[i]) >= 0 && i > minmax {
			continue // ignore P[i] below or on the upper line
		}

		for len(hull.List) > bot { // at least 2 points on the upper stack
			// test if  P[i] is left of the line at the stack top
			if el.hullCcw(hull.List[len(hull.List)-2], hull.List[len(hull.List)-1], P[i]) > 0 {
				break // P[i] is a new hull  vertex
			}

			hull.List = hull.List[:len(hull.List)-1] // pop top point off stack
		}
		hull.List = append(hull.List, P[i]) // push P[i] onto stack
	}
	if minmax != minmin {
		hull.List = append(hull.List, P[minmin]) // push  joining endpoint onto stack
	}

	return hull
}

func (el *PointList) ConvertToConcaveHull(n float64) (err error) {
	var hull PointList

	hull, err = el.ConcaveHull(n)
	if err != nil {
		return
	}

	el.List = hull.List
	return
}

func (e *PointList) equality(pointA, pointB Node) bool {
	return pointA.Loc[0] == pointB.Loc[0] && pointA.Loc[1] == pointB.Loc[1]
}

func (e *PointList) copyFrom(pointA, pointB *Node) {
	pointA.Id = pointB.Id
	pointA.Loc = pointB.Loc
	pointA.Rad = pointB.Rad
	pointA.Tag = pointB.Tag
}

func (e *PointList) decisionDistance(points []Node) float64 {
	var i int
	var curDistance float64
	var dst = e.pythagoras(points[0])
	for i = 1; i < len(points); i += 1 {
		curDistance = e.pythagoras(points[i])
		if curDistance < dst {
			dst = curDistance
		}
	}

	return dst
}

func (e *PointList) isContainedInTheList(pointA Node, points []Node) bool {
	for _, point := range points {
		if e.equality(pointA, point) {
			return true
		}
	}

	return false
}

func (e *PointList) pythagoras(point Node) float64 {
	return math.Sqrt(e.distanceSquared(point))
}

func (e *PointList) distanceSquared(point Node) float64 {
	return (math.Abs(point.Loc[0])-math.Abs(e.Loc[0]))*(math.Abs(point.Loc[0])-math.Abs(e.Loc[0])) +
		(math.Abs(point.Loc[1])-math.Abs(e.Loc[1]))*(math.Abs(point.Loc[1])-math.Abs(e.Loc[1]))
}

func (e *PointList) add(pointB Node) (node Node, err error) {
	err = node.SetLngLatDegrees(e.Loc[0]+pointB.Loc[0], e.Loc[1]+pointB.Loc[1])
	return
}

func (e *PointList) sub(pointB Node) (node Node, err error) {
	err = node.SetLngLatDegrees(e.Loc[0]-pointB.Loc[0], e.Loc[1]-pointB.Loc[1])
	return
}

func (e *PointList) plus(value float64) (node Node, err error) {
	err = node.SetLngLatDegrees(e.Loc[0]*value, e.Loc[1]*value)
	return
}

func (e *PointList) div(value float64) (node Node, err error) {
	err = node.SetLngLatDegrees(e.Loc[0]/value, e.Loc[1]/value)
	return
}

func (e *PointList) dotProduct(pointB Node) float64 {
	return e.Loc[0]*pointB.Loc[0] + e.Loc[1]*pointB.Loc[1]
}

func (e *PointList) distance(pointA, pointB Node) (distance float64, err error) {
	var l2 = pointA.distanceSquared(pointB)
	if l2 == 0.0 {
		distance = e.pythagoras(pointA) // v == w case
		return
	}

	// Consider the line extending the segment, parameterized as v + t (w - v)
	// We find projection of point p onto the line.
	// It falls where t = [(p-v) . (w-v)] / |w-v|^2
	var pA, pB, pC Node
	pA, err = e.sub(pointA)
	if err != nil {
		return
	}

	pB, err = pointB.sub(pointA)
	if err != nil {
		return
	}

	var t = pA.dotProduct(pB) / l2
	if t < 0.0 {
		distance = e.pythagoras(pointA)
		return
	} else if t > 1.0 {
		distance = e.pythagoras(pointB)
		return
	}
	pC, err = pointB.sub(pointA)
	if err != nil {
		return
	}

	pC, err = pC.plus(t)
	if err != nil {
		return
	}

	pC, err = pointA.add(pC)
	if err != nil {
		return
	}

	distance = e.pythagoras(pC)
	return
}

func (el *PointList) ConcaveHull(n float64) (hull PointList, err error) {
	hull = el.ConvexHull()

	var i, z int
	var eh, dd float64
	var found, intersects bool
	var ci1, ci2, pk Node
	var tmp = make([]Node, 0)

	var d, dTmp float64
	var skip bool
	var distance = 0.0

	for i = 0; i < len(hull.List)-1; i += 1 {
		// Find the nearest inner point pk ∈ G from the edge (ci1, ci2);
		ci1.copyFrom(hull.List[i])
		ci2.copyFrom(hull.List[i+1])

		distance = 0.0
		found = false

		for _, p := range el.List {
			// Skip points that are already in he hull
			if p.isContainedInTheList(hull.List) {
				continue
			}

			d, err = p.distance(ci1, ci2)
			if err != nil {
				return
			}
			skip = false
			for z = 0; !skip && z < len(hull.List)-1; z += 1 {
				dTmp, err = p.distance(hull.List[z], hull.List[z+1])
				if err != nil {
					return
				}
				skip = skip || dTmp < d
			}
			if skip {
				continue
			}

			if !found || distance > d {
				pk = p
				distance = d
				found = true
			}
		}

		if !found || pk.isContainedInTheList(hull.List) {
			continue
		}

		eh = ci1.pythagoras(ci2) // the length of the edge
		tmp = make([]Node, 0)
		tmp = append(tmp, ci1)
		tmp = append(tmp, ci2)

		dd = pk.decisionDistance(tmp)

		if eh/dd > n {
			// Check that new candidate edge will not intersect existing edges.
			intersects = el.hullCheckEdgeIntersectionList(hull.List, ci1, ci2, ci1, pk)
			intersects = intersects || el.hullCheckEdgeIntersectionList(hull.List, ci1, ci2, pk, ci2)
			if !intersects {
				hull.List = append(hull.List[:(i+1)], append([]Node{pk}, hull.List[(i+1):]...)...)
				i -= 1
			}
		}
	}

	return
}

func (el *PointList) hullCcw(p1, p2, p3 Node) float64 {
	return (p2.Rad[0]-p1.Rad[0])*(p3.Rad[1]-p1.Rad[1]) - (p2.Rad[1]-p1.Rad[1])*(p3.Rad[0]-p1.Rad[0])
}
