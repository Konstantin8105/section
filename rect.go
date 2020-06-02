package section

import "fmt"

// Rectangle
//
//	 |----- h --------|
//	 ****************** thk
//
type Rectangle struct {
	double height    //h
	double thickness //thk
}

func (r Rectangle) Geo(prec float64) string {
	var geo string
	geo += fmt.Sprintf("h   = %.5f;\n", h())
	geo += fmt.Sprintf("thk = %.5f;\n", thk())
	geo += fmt.Sprintf("Lc = %.5f;\n", prec)

	geo += `
	Point(000) = {+0.0000,+0.0000,+0.0000,Lc};
	Point(001) = {thk    ,+0.0000,+0.0000,Lc};
	Point(002) = {+0.0000,h      ,+0.0000,Lc};
	Point(003) = {thk    ,h      ,+0.0000,Lc};
	Line(1) = {1, 3};
	Line(2) = {0, 2};
	Line(3) = {0, 1};
	Line(4) = {2, 3};
	Line Loop(5) = {1, -4, -2, 3};
	Plane Surface(6) = {5};
`
	return geo
}
