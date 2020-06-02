package section

import "fmt"

////////////////////////////////////
////////////////////////////////////
////////// Shape: ANGLE ////////////
////////////////////////////////////
////////////////////////////////////
//    SCHEMA
//
//  --*r2
//  | *
//  | *thk
//  b *
//  | *r1
//  --*********r2
//    |--b----|
//
//
//
type Angle struct {
	Name      string
	Width     float64 //b
	Thickness float64 //thk
	Radius1   float64 //r1
	Radius2   float64 //r2
}

var Angles = []Angle{
	{"L50x5", 0.050, 0.005, 0.007, 0.0035},
	{"L60x6", 0.060, 0.006, 0.008, 0.0040},
	{"L63x6", 0.063, 0.006, 0.007, 0.0040},
	{"L63x6", 0.063, 0.006, 0.008, 0.0040},
	{"L70x7", 0.070, 0.007, 0.009, 0.0045},
	{"L75x7", 0.075, 0.007, 0.008, 0.0045},
	{"L75x8", 0.075, 0.008, 0.009, 0.0045},
	{"L80x8", 0.080, 0.008, 0.010, 0.0050},
	{"L90x9", 0.090, 0.009, 0.011, 0.0055},
	{"L100x10", 0.100, 0.010, 0.012, 0.0060},
	{"L120x12", 0.120, 0.012, 0.013, 0.0065},
	{"L150x15", 0.150, 0.015, 0.016, 0.0080},
}

func (a Angle) Geo(prec float64) string {
	var geo string

	geo += fmt.Sprintf("b = %.5f;\n", b())
	geo += fmt.Sprintf("thk = %.5f;\n", thk())
	geo += fmt.Sprintf("r1 = %.5f;\n", r1())
	geo += fmt.Sprintf("r2 = %.5f;\n", r2())
	geo += fmt.Sprintf("Lc = %.5f;\n", thk()/2)
	geo += fmt.Sprintf("Lc2 = %.5f;\n", prec) //h);

	geo += `
    Point(000) = {+0.0000,+0.0000,+0.0000,Lc};
    Point(001) = {b      ,+0.0000,+0.0000,Lc};
    Point(002) = {b      ,thk-r2 ,+0.0000,Lc2};
    Point(003) = {b-r2   ,thk-r2 ,+0.0000,Lc};
    Point(004) = {b-r2   ,thk    ,+0.0000,Lc2};
    
    Point(011) = {+0.0000,b      ,+0.0000,Lc};
    Point(012) = {thk-r2 ,b      ,+0.0000,Lc2};
    Point(013) = {thk-r2 ,b-r2   ,+0.0000,Lc};
    Point(014) = {thk    ,b-r2   ,+0.0000,Lc2};
    
    Point(020) = {thk+r1 ,thk+r1 ,+0.0000,Lc};
    Point(021) = {thk+r1 ,thk+00 ,+0.0000,Lc2};
    Point(022) = {thk+00 ,thk+r1 ,+0.0000,Lc2};
    Point(023) = {thk+00 ,thk+00 ,+0.0000,Lc};
    
    Line(1) = {0, 1};
    Line(2) = {1, 2};
    Line(3) = {4, 21};
    Line(4) = {22, 14};
    Line(5) = {0, 11};
    Line(6) = {11, 12};
    Circle(7) = {12, 13, 14};
    Circle(8) = {22, 20, 21};
    Circle(9) = {4, 3, 2};
    Line(10) = {22, 23};
    Line(11) = {23, 21};
    Line(12) = {4, 3};
    Line(13) = {3, 2};
    Line(14) = {12, 13};
    Line(15) = {13, 14};
    Line Loop(16) = {7, -15, -14};
    Plane Surface(17) = {16};
    Line Loop(18) = {14, 15, -4, 10, 11, -3, 12, 13, -2, -1, 5, 6};
    Plane Surface(19) = {18};
    Line Loop(20) = {8, -11, -10};
    Plane Surface(21) = {20};
    Line Loop(22) = {9, -13, -12};
    Plane Surface(23) = {22};
   `
	return geo
}
