package section

import "fmt"

// WPG - welded I-section
//	  |--b--|
//	  |     |
//	--******* tf
//	|    *
//	|    *
//	h    tw
//	|    *
//	|    *
//	|    *
//	--*******
type WPG struct {
	Name   string  // name of section
	Height float64 // height of I-section
	Width  float64 //b
	Tf     float64 //tf
	Tw     float64 //tw
	R      float64 // radius
}

func (w WPG) Geo(prec float64) string {
	var geo string
	geo += fmt.Sprintf("h = %.5f;\n", w.Height)
	geo += fmt.Sprintf("b = %.5f;\n", w.Width)
	geo += fmt.Sprintf("s = %.5f;\n", w.Tw)
	geo += fmt.Sprintf("t = %.5f;\n", w.Tf)
	geo += fmt.Sprintf("R = %.5f;\n", w.R)
	geo += fmt.Sprintf("prec = %.5f;\n", prec)
	geo += `
    Point(0) ={0,0,0,prec};
    Point(1) ={b,0,0,prec};
    Point(2) ={b,t,0,prec};
    Point(3) ={b/2+s/2+R,t,0,prec};
    Point(5) ={b/2+s/2,t+R,0,prec};
    Point(6) ={b/2+s/2,h-t-R,0,prec};
    Point(8) ={b/2+s/2+R,h-t,0,prec};
    Point(9) ={b,h-t,0,prec};
    Point(10)={b,h,0,prec};
    Point(11)={0,h,0,prec};
    Point(12)={0,h-t,0,prec};
    Point(13)={b/2-s/2-R,h-t,0,prec};
    Point(14)={b/2-s/2,h-t-R,0,prec};
    Point(16)={b/2-s/2,t+R,0,prec};
    Point(18)={b/2-s/2-R,t,0,prec};
    Point(19)={0,t,0,prec};
    Point(20)={b/2+s/2,h/2,0,prec};
    Point(21)={b/2-s/2,h/2,0,prec};
    Point(22)={b/2-s/2,t,0,prec};
    Point(23)={b/2+s/2,t,0,prec};
    Point(24)={b/2-s/2,h-t,0,prec};
    Point(25)={b/2+s/2,h-t,0,prec};
    Line(1)  = {0, 1};
    Line(2)  = {1, 2};
    Line(3)  = {2, 3};
    Line(4)  = {0, 19};
    Line(5)  = {19, 18};
    Line(6)  = {8, 9};
    Line(7)  = {9, 10};
    Line(8)  = {10, 11};
    Line(9)  = {11, 12};
    Line(10) = {12, 13};
    Line(11) = {6, 20};
    Line(12) = {20, 5};
    Line(13) = {16, 21};
    Line(14) = {21, 14};
    Line(19) = {18, 22};
    Line(20) = {22, 16};
    Line(21) = {3, 23};
    Line(22) = {23, 5};
    Line(23) = {13, 24};
    Line(24) = {24, 14};
    Line(25) = {6, 25};
    Line(26) = {25, 8};
    Line Loop(27) = {1, 2, 3, 21, 22, -12, -11, 25, 26, 6, 7, 8, 9, 10, 23, 24, -14, -13, -20, -19, -5, -4};
    Plane Surface(28) = {27};
    `
	return geo
}
