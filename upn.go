package section

import "fmt"

// UPN
//	 --*********r2
//	 | *r1   tf
//	 | *
//	 | *
//	 | *
//	 h *tw
//	 | *
//	 | *
//	 | *
//	 | *r1   tf
//	 --*********r2
//	   |--b----|
type UPN struct {
	Name    string
	width   float64 //b
	height  float64 //h
	ThkWeb  float64 //tw
	ThkFl   float64 //tf
	radius1 float64 //r1
	radius2 float64 //r2
}

var UPNs = []UPN{
	{"UPN120", 0.1200, 0.0550, 0.0090, 0.0070, 0.0090, 0.0045},
	{"UPN140", 0.1400, 0.0600, 0.0100, 0.0070, 0.0100, 0.0050},
	{"UPN160", 0.1600, 0.0650, 0.0105, 0.0075, 0.0105, 0.0055},
	{"UPN180", 0.1800, 0.0700, 0.0110, 0.0080, 0.0110, 0.0055},
	{"UPN200", 0.2000, 0.0750, 0.0115, 0.0085, 0.0115, 0.0060},
	{"UPN240", 0.2400, 0.0850, 0.0130, 0.0095, 0.0130, 0.0065},
	{"UPN300", 0.3000, 0.1000, 0.0160, 0.0100, 0.0160, 0.0080},
	{"UPN400", 0.4000, 0.1100, 0.0180, 0.0140, 0.0180, 0.0090},
}

func (u UPN) Geo(prec float64) string {
	var geo string
	geo += fmt.Sprintf("h  = %.5f;\n", h())
	geo += fmt.Sprintf("b  = %.5f;\n", b())
	geo += fmt.Sprintf("tf = %.5f;\n", tf())
	geo += fmt.Sprintf("tw = %.5f;\n", tw())
	geo += fmt.Sprintf("r1 = %.5f;\n", r1())
	geo += fmt.Sprintf("r2 = %.5f;\n", r2())
	geo += fmt.Sprintf("Lc = %.5f;\n", tw()/2)
	geo += fmt.Sprintf("Lc2 = %.5f;\n", prec)
	geo += `
	u = b/2;
	degree = 0.08;
	If(h > 0.300)
	u = (b-tw)/2;
	degree = 0.05;
	EndIf
	
	Point(000) = {+0.0000,+0.0000,+0.0000,Lc};
	Point(002) = {b      ,+0.0000,+0.0000,Lc};
	Point(004) = {b-u    ,0+tf   ,+0.0000,Lc};
	Point(005) = {+0.0000,h/2    ,+0.0000,Lc};
	Point(006) = {tw     ,h/2    ,+0.0000,Lc};
	
	
	Point(100) = {+0.0000,h      ,+0.0000,Lc};
	Point(102) = {b      ,h      ,+0.0000,Lc};
	Point(104) = {b-u    ,h-tf   ,+0.0000,Lc};
	
	betta      = (-Atan(degree) + 3.1415926/2)/2;
	Yc         = tf - u*degree;
	Point(007) = {b      ,Yc     ,+0.0000,Lc};
	Point(008) = {b      ,Yc-r2*Tan(betta),+0.0000,Lc2};
	Point(009) = {b-r2   ,Yc-r2*Tan(betta),+0.0000,Lc2};
	Point(010) = {b-r2+r2*Cos(2*betta),Yc-r2*Tan(betta)+r2*Sin(2*betta),+0.0000,Lc2};
	Point(107) = {b      ,h-Yc   ,+0.0000,Lc};
	Point(108) = {b      ,h-(Yc-r2*Tan(betta)),+0.0000,Lc2};
	Point(109) = {b-r2   ,h-(Yc-r2*Tan(betta)),+0.0000,Lc2};
	Point(110) = {b-r2+r2*Cos(2*betta),h-(Yc-r2*Tan(betta)+r2*Sin(2*betta)),+0.0000,Lc2};
	
	Yc         = tf + (b-tw-u)*degree;
	Point(027) = {tw     ,Yc              ,+0.0000,Lc2};
	Point(028) = {tw     ,Yc+r1*Tan(betta),+0.0000,Lc2};
	Point(029) = {tw+r1  ,Yc+r1*Tan(betta),+0.0000,Lc};
	Point(030) = {tw+r1-r1*Cos(2*betta),Yc+r1*Tan(betta)-r1*Sin(2*betta),+0.0000,Lc2};
	Point(127) = {tw     ,h-Yc                ,+0.0000,Lc2};
	
	Point(128) = {tw     ,h-(Yc+r1*Tan(betta)),+0.0000,Lc2};
	Point(129) = {tw+r1  ,h-(Yc+r1*Tan(betta)),+0.0000,Lc};
	Point(130) = {tw+r1-r1*Cos(2*betta),h-(Yc+r1*Tan(betta)-r1*Sin(2*betta)),+0.0000,Lc2};
	
	Line(1) = {102, 100};
	Line(2) = {100, 5};
	Line(3) = {5, 0};
	Line(4) = {0, 2};
	Line(5) = {2, 8};
	Line(6) = {10, 4};
	Line(7) = {4, 30};
	Line(8) = {28, 6};
	Line(9) = {6, 128};
	Line(10) = {130, 104};
	Line(11) = {104, 110};
	Line(12) = {108, 102};
	Line(13) = {109, 110};
	Line(14) = {109, 108};
	Line(15) = {127, 128};
	Line(16) = {127, 130};
	Line(17) = {28, 27};
	Line(18) = {27, 30};
	Line(19) = {10, 9};
	Line(20) = {9, 8};
	Circle(21) = {10, 9, 8};
	Circle(24) = {110, 109, 108};
	
	Line Loop(25) = {14, -24, -13};
	Plane Surface(26) = {25};
	Line Loop(27) = {19, 20, -21};
	Plane Surface(28) = {27};
	Line Loop(29) = {11, -13, 14, 12, 1, 2, 3, 4, 5, -20, -19, 6, 7, -18, -17, 8, 9, -15, 16, 10};
	Plane Surface(30) = {29};
	
	Circle(31) = {128, 129, 130};
	Line Loop(32) = {31, -16, 15};
	Plane Surface(33) = {32};
	Circle(34) = {28, 29, 30};
	Line Loop(35) = {18, -34, 17};
	Plane Surface(36) = {35};
`
	return geo
}
