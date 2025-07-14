package section

import (
	"bytes"
	"fmt"
	"sync"
	"text/tabwriter"
	"text/template"

	"github.com/Konstantin8105/efmt"
)

type Geor interface {
	Geo(prec float64) string
	GetName() string
}

var (
	mutex sync.Mutex
	ps    []Property
)

func GetProperty(g Geor) (p Property, err error) {
	// stored data
	for i := range ps {
		if ps[i].Name == g.GetName() {
			pc := ps[i] // copy
			return pc, nil
		}
	}
	pt, err := Calculate(g)
	if err != nil {
		return
	}
	p = *pt // copy
	// add to stored data
	mutex.Lock()
	defer func() {
		mutex.Unlock()
	}()
	ps = append(ps, p)
	return
}

func GetList() (list []Geor) {
	for i := range Angles {
		list = append(list, Angles[i])
	}
	for i := range Isections {
		list = append(list, Isections[i])
	}
	for i := range UPNs {
		list = append(list, UPNs[i])
	}
	for i := range Rectangles {
		list = append(list, Rectangles[i])
	}
	return
}

func Get(name string) (_ Geor, err error) {
	list := GetList()
	for i := range list {
		if name == list[i].GetName() {
			return list[i], nil
		}
	}
	return nil, fmt.Errorf("Section with name: `%s` is not found", name)
}

// //////////////////////////////////
// //////////////////////////////////
// //////// Shape: ANGLE ////////////
// //////////////////////////////////
// //////////////////////////////////
//
//	  SCHEMA
//
//	--*r2
//	| *
//	| *thk
//	b *
//	| *r1
//	--*********r2
//	  |--b----|
type Angle struct {
	Name    string
	Width   float64 //b
	Thk     float64 //Thickness
	Radius1 float64 //r1
	Radius2 float64 //r2
}

func (a Angle) GetName() string {
	if a.Name == "" {
		return fmt.Sprintf("L%.2fx%.2f",
			a.Width*1e3,
			a.Thk*1e3,
		)
	}
	return a.Name
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
	// TODO: use text/template
	var geo string

	geo += fmt.Sprintf("b = %.5f;\n", a.Width)
	geo += fmt.Sprintf("thk = %.5f;\n", a.Thk)
	geo += fmt.Sprintf("r1 = %.5f;\n", a.Radius1)
	geo += fmt.Sprintf("r2 = %.5f;\n", a.Radius2)
	geo += fmt.Sprintf("Lc = %.5f;\n", a.Thk/2)
	geo += fmt.Sprintf("Lc2 = %.5f;\n", prec)

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

// TODO: it is tube ???

// SCHEMA
//
//	  **   DIA
//	*    *
//	*    * THK
//	  **
type Cylinder struct {
	Name string
	Od   float64
	Thk  float64 //thk
}

func (c Cylinder) GetName() string {
	if c.Name == "" {
		return fmt.Sprintf("DIA%.2fx%.2f",
			1e3*c.Od,
			1e3*c.Thk,
		)
	}
	return c.Name
}

func (c Cylinder) Geo(prec float64) string {
	// TODO: use text/template
	var geo string
	geo += fmt.Sprintf("dia = %.5f;\n", c.Od)
	geo += fmt.Sprintf("thk = %.5f;\n", c.Thk)
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

////////////////////////////////////
////////////////////////////////////
////////// I - section  ////////////
////////////////////////////////////
////////////////////////////////////
//    SCHEMA
//
//    |--b--|
//    |     |
//  --******* tf
//  |   r*r
//  |    *
//  h    tw
//  |    *
//  |    *
//  |   r*r
//  --*******

type Isection struct {
	Name   string
	H      float64 // height
	B      float64 // width
	Tw     float64 //tw
	Tf     float64 //tf
	Radius float64 //r
}

func (i Isection) GetName() string {
	if i.Name == "" {
		return fmt.Sprintf("WPG H%.2f x B%.2f x Tf%.2f x Tw%.2f",
			1e3*i.H,
			1e3*i.B,
			1e3*i.Tf,
			1e3*i.Tw,
		)
	}
	return i.Name
}

var Isections = []Isection{
	///
	/// ASCM STO 20-93. PROFILE "B"
	///
	{"10B1-ASCM", 0.1000, 0.0550, 0.0041, 0.0057, 0.0070},
	{"12B1-ASCM", 0.1176, 0.0640, 0.0038, 0.0051, 0.0070},
	{"12B2-ASCM", 0.1200, 0.0640, 0.0044, 0.0063, 0.0070},
	{"14B1-ASCM", 0.1374, 0.0730, 0.0038, 0.0056, 0.0070},
	{"14B2-ASCM", 0.1400, 0.0730, 0.0047, 0.0069, 0.0070},
	{"16B1-ASCM", 0.1570, 0.0820, 0.0040, 0.0059, 0.0090},
	{"16B2-ASCM", 0.1600, 0.0820, 0.0050, 0.0074, 0.0090},
	{"18B1-ASCM", 0.1770, 0.0910, 0.0043, 0.0065, 0.0090},
	{"18B2-ASCM", 0.1800, 0.0910, 0.0053, 0.0080, 0.0090},
	{"20B1-ASCM", 0.2000, 0.1000, 0.0055, 0.0080, 0.0110},
	{"25B1-ASCM", 0.2480, 0.1240, 0.0050, 0.0080, 0.0120},
	{"25B2-ASCM", 0.2500, 0.1250, 0.0060, 0.0090, 0.0120},
	{"30B1-ASCM", 0.2980, 0.1490, 0.0055, 0.0080, 0.0130},
	{"30B2-ASCM", 0.3000, 0.1500, 0.0065, 0.0090, 0.0130},
	{"35B1-ASCM", 0.3460, 0.1740, 0.0060, 0.0090, 0.0140},
	{"35B2-ASCM", 0.3500, 0.1750, 0.0070, 0.0110, 0.0140},
	{"40B1-ASCM", 0.3960, 0.1990, 0.0070, 0.0110, 0.0160},
	{"40B2-ASCM", 0.4000, 0.2000, 0.0080, 0.0130, 0.0160},

	///
	/// ASCM STO 20-93. PROFILE "K"
	///
	{"20K2-ASCM", 0.2000, 0.2000, 0.0080, 0.0120, 0.0130},
	{"25K2-ASCM", 0.2500, 0.2500, 0.0090, 0.0140, 0.0160},
	{"30K2-ASCM", 0.3000, 0.3000, 0.0100, 0.0150, 0.0180},
	{"35K2-ASCM", 0.3500, 0.3500, 0.0120, 0.0190, 0.0200},
	{"40K2-ASCM", 0.4000, 0.4000, 0.0130, 0.0210, 0.0220},

	///
	/// ASCM STO 20-93. PROFILE "SH"
	///
	{"100SH1-ASCM", 0.9900, 0.3200, 0.0160, 0.0210, 0.0300},

	///
	/// European code. PROFILE "IPE"
	///
	{"IPE160", 0.1600, 0.0820, 0.0050, 0.0074, 0.0090},
	{"IPE180", 0.1800, 0.0910, 0.0053, 0.0080, 0.0090},
	{"IPE200", 0.2000, 0.1000, 0.0056, 0.0085, 0.0120},
	{"IPE240", 0.2400, 0.1200, 0.0062, 0.0098, 0.0150},
	{"IPE300", 0.3000, 0.1500, 0.0071, 0.0107, 0.0150},
	{"IPE500", 0.5000, 0.2000, 0.0102, 0.0160, 0.0370},
	{"IPE550", 0.5500, 0.2100, 0.0111, 0.0172, 0.0412},

	///
	/// European code. PROFILE "HEA"
	///
	{"HEA140", 0.1330, 0.1400, 0.0055, 0.0085, 0.0120},
	{"HEA160", 0.1520, 0.1600, 0.0060, 0.0090, 0.0150},
	{"HEA180", 0.1710, 0.1800, 0.0060, 0.0095, 0.0150},
	{"HEA200", 0.1900, 0.2000, 0.0065, 0.0100, 0.0180},
	{"HEA240", 0.2300, 0.2400, 0.0075, 0.0120, 0.0210},
	{"HEA300", 0.2900, 0.3000, 0.0085, 0.0140, 0.0270},
	{"HEA500", 0.4900, 0.3000, 0.0120, 0.0230, 0.0270},

	///
	/// European code. PROFILE "HEB"
	///
	{"HEB140", 0.1400, 0.1400, 0.0070, 0.0120, 0.0120},
	{"HEB160", 0.1600, 0.1600, 0.0080, 0.0130, 0.0150},
	{"HEB180", 0.1800, 0.1800, 0.0085, 0.0140, 0.0150},
	{"HEB200", 0.2000, 0.2000, 0.0090, 0.0150, 0.0180},
	{"HEB240", 0.2400, 0.2400, 0.0100, 0.0170, 0.0210},
	{"HEB300", 0.3000, 0.3000, 0.0110, 0.0190, 0.0270},
}

func (is Isection) Geo(prec float64) string {
	tmplString := `
h        = {{ .H }}  ;
b        = {{ .B }}  ;
s        = {{ .Tw }} ;
t        = {{ .Tf }} ;
R        = {{ .Radius }};
prec     = {{ .Prec }} ;
prec2    = {{ .Tw }} ;
Point(0) ={0,0,0,prec2};
Point(1) ={b,0,0,prec2};
Point(2) ={b,t,0,prec2};
Point(3) ={b/2+s/2+R,t,0,prec};
Point(4) ={b/2+s/2+R,t+R,0,prec};
Point(5) ={b/2+s/2,t+R,0,prec};
Point(6) ={b/2+s/2,h-t-R,0,prec};
Point(7) ={b/2+s/2+R,h-t-R,0,prec};
Point(8) ={b/2+s/2+R,h-t,0,prec};
Point(9) ={b,h-t,0,prec2};
Point(10)={b,h,0,prec2};
Point(11)={0,h,0,prec2};
Point(12)={0,h-t,0,prec2};
Point(13)={b/2-s/2-R,h-t,0,prec};
Point(14)={b/2-s/2,h-t-R,0,prec};
Point(15)={b/2-s/2-R,h-t-R,0,prec};
Point(16)={b/2-s/2,t+R,0,prec};
Point(17)={b/2-s/2-R,t+R,0,0,prec};
Point(18)={b/2-s/2-R,t,0,prec};
Point(19)={0,t,0,prec2};
Point(20)={b/2+s/2,h/2,0,prec2};
Point(21)={b/2-s/2,h/2,0,prec2};
Point(22)={b/2-s/2,t,0,prec2};
Point(23)={b/2+s/2,t,0,prec2};
Point(24)={b/2-s/2,h-t,0,prec2};
Point(25)={b/2+s/2,h-t,0,prec2};
Line(1) = {0, 1};
Line(2) = {1, 2};
Line(3) = {2, 3};
Line(4) = {0, 19};
Line(5) = {19, 18};
Line(6) = {8, 9};
Line(7) = {9, 10};
Line(8) = {10, 11};
Line(9) = {11, 12};
Line(10) = {12, 13};
Line(11) = {6, 20};
Line(12) = {20, 5};
Line(13) = {16, 21};
Line(14) = {21, 14};
Circle(15) = {6, 7, 8};
Circle(16) = {14, 15, 13};
Circle(17) = {18, 17, 16};
Circle(18) = {3, 4, 5};
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
Line Loop(29) = {16, 23, 24};
Plane Surface(30) = {29};
Line Loop(31) = {25, 26, -15};
Plane Surface(32) = {31};
Line Loop(33) = {20, -17, 19};
Plane Surface(34) = {33};
Line Loop(35) = {18, -22, -21};
Plane Surface(36) = {35};
`
	var (
		value = struct {
			Isection
			Prec float64
		}{Isection: is, Prec: prec}
		tmpl = template.Must(template.New("geo").Parse(tmplString))
		buf  bytes.Buffer
	)
	if err := tmpl.Execute(&buf, value); err != nil {
		panic(err)
	}
	return buf.String()
}

// Rectangle
//
//	--*
//	| *
//	h *
//	| *
//	--*
//	  thk
type Rectangle struct {
	Name string
	H    float64 //height
	Thk  float64 //thickness
}

func (r Rectangle) GetName() string {
	if r.Name == "" {
		return fmt.Sprintf("Rectangle H%.2f x Thk%.2f",
			1e3*r.H,
			1e3*r.Thk,
		)
	}
	return r.Name
}

func (r Rectangle) Geo(prec float64) string {
	// TODO: use text/template
	var geo string
	geo += fmt.Sprintf("h   = %.5f;\n", r.H)
	geo += fmt.Sprintf("thk = %.5f;\n", r.Thk)
	geo += fmt.Sprintf("Lc = %.5f;\n", prec)

	geo += `
	Point(000) = {-thk/2.0,+0.0000,+0.0000,Lc};
	Point(001) = {+thk/2.0,+0.0000,+0.0000,Lc};
	Point(002) = {-thk/2.0,h      ,+0.0000,Lc};
	Point(003) = {+thk/2.0,h      ,+0.0000,Lc};
	Line(1) = {1, 3};
	Line(2) = {0, 2};
	Line(3) = {0, 1};
	Line(4) = {2, 3};
	Line Loop(5) = {1, -4, -2, 3};
	Plane Surface(6) = {5};
`
	return geo
}

var Rectangles = []Rectangle{
	{"Plate 50x5", 0.050, 0.005},
	{"Plate 60x6", 0.060, 0.006},
	{"Plate 75x7", 0.075, 0.007},
	{"Plate 100x10", 0.100, 0.010},
}

// Plate
type Plate struct {
	Xc, Yc float64 // location of center mass of plate
	X, Y   float64 // sizes by directions x and y
}

type PlateGroup struct {
	Name   string
	Plates []Plate
}

func (pg PlateGroup) GetName() string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 3, ' ', tabwriter.TabIndent)
	if pg.Name != "" {
		fmt.Fprintf(w, "%s\n", pg.Name)
	} else {
		fmt.Fprintf(w, "%s\n", "Plate group")
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "№\tXc\tYc\tX\tY\n")
	for i, p := range pg.Plates {
		fmt.Fprintf(w,
			"%d\t%s\t%s\t%s\t%s\n",
			i,
			efmt.Sprint(p.Xc), efmt.Sprint(p.Yc),
			efmt.Sprint(p.X), efmt.Sprint(p.Y),
		)
	}
	fmt.Fprintf(w, "\n")
	w.Flush()
	return buf.String()
}

func (pg PlateGroup) Geo(prec float64) string {
	var geo string
	geo += fmt.Sprintf("Lc = %.5f;\n", prec)
	for i, p := range pg.Plates {
		geo += fmt.Sprintf("Xc = %.5f;\n", p.Xc)
		geo += fmt.Sprintf("Yc = %.5f;\n", p.Yc)
		geo += fmt.Sprintf("X  = %.5f;\n", p.X)
		geo += fmt.Sprintf("Y  = %.5f;\n", p.Y)
		geo += fmt.Sprintf("ID = %d;\n", (i+1)*10000) // maximal 10000 plates
		geo += `
	Point(ID + 000) = {Xc - X/2.0 , Yc - Y/2.0,+0.0000,Lc};
	Point(ID + 001) = {Xc + X/2.0 , Yc - Y/2.0,+0.0000,Lc};
	Point(ID + 002) = {Xc - X/2.0 , Yc + Y/2.0,+0.0000,Lc};
	Point(ID + 003) = {Xc + X/2.0 , Yc + Y/2.0,+0.0000,Lc};
	Line(ID + 1) = { ID + 1, ID + 3};
	Line(ID + 2) = { ID + 0, ID + 2};
	Line(ID + 3) = { ID + 0, ID + 1};
	Line(ID + 4) = { ID + 2, ID + 3};
	Line Loop(ID + 5) = {ID + 1, -(ID + 4),-( ID + 2), ID + 3};
	Plane Surface(ID + 6) = {ID + 5};
`
	}
	return geo
}

// Tsection
//
//	      Thk
//	       * -
//	       * |
//	       * H
//	       * |
//	       * -
//	************** Thk2
//	|----- L ----|
type Tsection struct {
	Name string

	H   float64 // height
	Thk float64 // thickness

	L    float64 // length
	Thk2 float64 // thickness
}

func (t Tsection) GetName() string {
	if t.Name == "" {
		return fmt.Sprintf("Tsection H%.2f x L%.2f x Thk%.2f x Thk2%.2f",
			1e3*t.H,
			1e3*t.L,
			1e3*t.Thk,
			1e3*t.Thk2,
		)
	}
	return t.Name
}

func (t Tsection) Geo(prec float64) string {
	// TODO: use text/template
	var geo string
	geo += fmt.Sprintf("H    = %.5f;\n", t.H)
	geo += fmt.Sprintf("Thk  = %.5f;\n", t.Thk)
	geo += fmt.Sprintf("L    = %.5f;\n", t.L)
	geo += fmt.Sprintf("Thk2 = %.5f;\n", t.Thk2)
	geo += fmt.Sprintf("Lc   = %.5f;\n", prec)

	geo += `
	Point(000) = {-Thk/2.0,+0.0000,+0.0000,Lc};
	Point(001) = {-Thk/2.0,H      ,+0.0000,Lc};
	Point(002) = {+Thk/2.0,H      ,+0.0000,Lc};
	Point(003) = {+Thk/2.0,+0.0000,+0.0000,Lc};
	Point(004) = {+L/2.0  ,+0.0000,+0.0000,Lc};
	Point(005) = {+L/2.0  ,-Thk2/2.0,+0.0000,Lc};
	Point(006) = {-L/2.0  ,-Thk2/2.0,+0.0000,Lc};
	Point(007) = {-L/2.0  ,+0.0000,+0.0000,Lc};
	Line(1) = {0, 1};
	Line(2) = {1, 2};
	Line(3) = {2, 3};
	Line(4) = {3, 4};
	Line(5) = {4, 5};
	Line(6) = {5, 6};
	Line(7) = {6, 7};
	Line(8) = {7, 0};
	Line Loop(10) = {1,2,3,4,5,6,7,8};
	Plane Surface(20) = {10};
`
	return geo
}

// UPN
//
//	--*********r2
//	| *r1   tf
//	| *
//	| *
//	| *
//	h *tw
//	| *
//	| *
//	| *
//	| *r1   tf
//	--*********r2
//	  |--b----|
type UPN struct {
	Name    string
	H       float64 //h height
	B       float64 //b  Width
	Tf      float64 //tf
	Tw      float64 //tw
	Radius1 float64 //r1
	Radius2 float64 //r2

	// TODO: angle of flange
}

func (u UPN) GetName() string {
	if u.Name == "" {
		return fmt.Sprintf("UPN H%.2f x B%.2f x Tf%.2f x Tw%.2f",
			1e3*u.H,
			1e3*u.B,
			1e3*u.Tf,
			1e3*u.Tw,
		)
	}
	return u.Name
}

func (u UPN) String() string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', tabwriter.TabIndent)
	fmt.Fprintf(w, "Name\t| \t| %s\n", u.Name)
	fmt.Fprintf(w, "Height\t| h\t| %10.1f mm\n", u.H*1000)
	fmt.Fprintf(w, "Width\t| b\t| %10.1f mm\n", u.B*1000)
	fmt.Fprintf(w, "Thickness of flange\t| tf\t| %10.1f mm\n", u.Tf*1000)
	fmt.Fprintf(w, "Thickness of wall\t| tw\t| %10.1f mm\n", u.Tw*1000)
	fmt.Fprintf(w, "Radius1\t| r1\t| %10.1f mm\n", u.Radius1*1000)
	fmt.Fprintf(w, "Radius2\t| r2\t| %10.1f mm\n", u.Radius2*1000)
	w.Flush()
	return buf.String()
}

var UPNs = []UPN{ // TODO: check
	{"UPN120 DIN 1025-5-1994", 0.1200, 0.0550, 0.0090, 0.0070, 0.0090, 0.0045},
	{"UPN140 DIN 1025-5-1994", 0.1400, 0.0600, 0.0100, 0.0070, 0.0100, 0.0050},
	{"UPN160 DIN 1025-5-1994", 0.1600, 0.0650, 0.0105, 0.0075, 0.0105, 0.0055},
	{"UPN180 DIN 1025-5-1994", 0.1800, 0.0700, 0.0110, 0.0080, 0.0110, 0.0055},
	{"UPN200 DIN 1025-5-1994", 0.2000, 0.0750, 0.0115, 0.0085, 0.0115, 0.0060},
	{"UPN240 DIN 1025-5-1994", 0.2400, 0.0850, 0.0130, 0.0095, 0.0130, 0.0065},
	{"UPN300 DIN 1025-5-1994", 0.3000, 0.1000, 0.0160, 0.0100, 0.0160, 0.0080},
	{"UPN400 DIN 1025-5-1994", 0.4000, 0.1100, 0.0180, 0.0140, 0.0180, 0.0090},

	{"Швеллер 12У ГОСТ 8240", 0.120, 0.052, 0.0078, 0.0048, 0.0075, 0.0030},
	{"Швеллер 16У ГОСТ 8240", 0.160, 0.064, 0.0084, 0.0050, 0.0085, 0.0035},
	{"Швеллер 20У ГОСТ 8240", 0.200, 0.076, 0.0090, 0.0052, 0.0095, 0.0040},
	{"Швеллер 24У ГОСТ 8240", 0.240, 0.090, 0.0100, 0.0056, 0.0105, 0.0040},
	{"Швеллер 30У ГОСТ 8240", 0.300, 0.100, 0.0110, 0.0065, 0.0120, 0.0050},
	{"Швеллер 36У ГОСТ 8240", 0.360, 0.110, 0.0126, 0.0075, 0.0140, 0.0060},
	{"Швеллер 40У ГОСТ 8240", 0.400, 0.115, 0.0135, 0.0080, 0.0150, 0.0060},
}

func init() {
	// initialize double channels
	for _, upn := range UPNs {
		var is Isection
		is.Name = upn.Name + ",Double"
		is.H = upn.H
		is.B = upn.B * 2.0
		is.Tf = upn.Tf
		is.Tw = upn.Tw * 2.0
		is.Radius = upn.Radius1
		Isections = append(Isections, is)
	}
}

func (u UPN) Geo(prec float64) string {
	// TODO: use text/template
	// TODO: simplification of geo file
	var geo string
	geo += fmt.Sprintf("h  = %.5f;\n", u.H)
	geo += fmt.Sprintf("b  = %.5f;\n", u.B)
	geo += fmt.Sprintf("tf = %.5f;\n", u.Tf)
	geo += fmt.Sprintf("tw = %.5f;\n", u.Tw)
	geo += fmt.Sprintf("r1 = %.5f;\n", u.Radius1)
	geo += fmt.Sprintf("r2 = %.5f;\n", u.Radius2)
	geo += fmt.Sprintf("Lc = %.5f;\n", u.Tw/2)
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
// type WPG struct {
// 	Name   string  // name of section
// 	Height float64 // height of I-section
// 	Width  float64 //b
// 	Tf     float64 //tf
// 	Tw     float64 //tw
// 	R      float64 // radius
// }
//
// func (w WPG) Geo(prec float64) string {
// 	var geo string
// 	geo += fmt.Sprintf("h = %.5f;\n", w.Height)
// 	geo += fmt.Sprintf("b = %.5f;\n", w.Width)
// 	geo += fmt.Sprintf("s = %.5f;\n", w.Tw)
// 	geo += fmt.Sprintf("t = %.5f;\n", w.Tf)
// 	geo += fmt.Sprintf("R = %.5f;\n", w.R)
// 	geo += fmt.Sprintf("prec = %.5f;\n", prec)
// 	geo += `
//     Point(0) ={0,0,0,prec};
//     Point(1) ={b,0,0,prec};
//     Point(2) ={b,t,0,prec};
//     Point(3) ={b/2+s/2+R,t,0,prec};
//     Point(5) ={b/2+s/2,t+R,0,prec};
//     Point(6) ={b/2+s/2,h-t-R,0,prec};
//     Point(8) ={b/2+s/2+R,h-t,0,prec};
//     Point(9) ={b,h-t,0,prec};
//     Point(10)={b,h,0,prec};
//     Point(11)={0,h,0,prec};
//     Point(12)={0,h-t,0,prec};
//     Point(13)={b/2-s/2-R,h-t,0,prec};
//     Point(14)={b/2-s/2,h-t-R,0,prec};
//     Point(16)={b/2-s/2,t+R,0,prec};
//     Point(18)={b/2-s/2-R,t,0,prec};
//     Point(19)={0,t,0,prec};
//     Point(20)={b/2+s/2,h/2,0,prec};
//     Point(21)={b/2-s/2,h/2,0,prec};
//     Point(22)={b/2-s/2,t,0,prec};
//     Point(23)={b/2+s/2,t,0,prec};
//     Point(24)={b/2-s/2,h-t,0,prec};
//     Point(25)={b/2+s/2,h-t,0,prec};
//     Line(1)  = {0, 1};
//     Line(2)  = {1, 2};
//     Line(3)  = {2, 3};
//     Line(4)  = {0, 19};
//     Line(5)  = {19, 18};
//     Line(6)  = {8, 9};
//     Line(7)  = {9, 10};
//     Line(8)  = {10, 11};
//     Line(9)  = {11, 12};
//     Line(10) = {12, 13};
//     Line(11) = {6, 20};
//     Line(12) = {20, 5};
//     Line(13) = {16, 21};
//     Line(14) = {21, 14};
//     Line(19) = {18, 22};
//     Line(20) = {22, 16};
//     Line(21) = {3, 23};
//     Line(22) = {23, 5};
//     Line(23) = {13, 24};
//     Line(24) = {24, 14};
//     Line(25) = {6, 25};
//     Line(26) = {25, 8};
//     Line Loop(27) = {1, 2, 3, 21, 22, -12, -11, 25, 26, 6, 7, 8, 9, 10, 23, 24, -14, -13, -20, -19, -5, -4};
//     Plane Surface(28) = {27};
//     `
// 	return geo
// }
