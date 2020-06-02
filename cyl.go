package section

import "fmt"

// TODO: it is tube ???

//    SCHEMA
//
//       **   DIA
//     *    *
//     *    * THK
//       **
//
type Cylinder struct {
	double Diameter
	double thickness //thk
}

func (c Cylinder) Geo(prec float64) string {
	var geo string
	geo += fmt.Sprintf("dia = %.5f;\n", diameter())
	geo += fmt.Sprintf("thk = %.5f;\n", thk())
	geo += fmt.Sprintf("Lc = %.5f;\n", prec)

	geo += `
	r = dia/2;
	Point(001) = {+0.0000,+0.0000,+0.0000,Lc};
	Point(002) = {+r,+0.0000,+0.0000,Lc};
	Point(003) = {-r,+0.0000,+0.0000,Lc};
	Point(004) = {+0.0000,+r,+0.0000,Lc};
	Point(005) = {+0.0000,-r,+0.0000,Lc};
	Point(012) = {+r-thk,+0.0000,+0.0000,Lc};
	Point(013) = {-r+thk,+0.0000,+0.0000,Lc};
	Point(014) = {+0.0000,+r-thk,+0.0000,Lc};
	Point(015) = {+0.0000,-r+thk,+0.0000,Lc};
	Circle(1) = {4, 1, 2};
	Circle(2) = {2, 1, 5};
	Circle(3) = {5, 1, 3};
	Circle(4) = {3, 1, 4};
	Circle(5) = {14, 1, 12};
	Circle(6) = {12, 1, 15};
	Circle(7) = {15, 1, 13};
	Circle(8) = {13, 1, 14};
	Line(9) = {12, 2};
	Line(10) = {13, 3};
	Line Loop(11) = {2, 3, -10, -7, -6, 9};
	Plane Surface(12) = {11};
	Line Loop(13) = {8, 5, 9, -1, -4, -10};
	Plane Surface(14) = {13};
`
	return geo
}
