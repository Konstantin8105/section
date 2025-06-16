package section

import (
	"bytes"
	"fmt"
	"math"
	"text/tabwriter"

	"github.com/Konstantin8105/efmt"
	"github.com/Konstantin8105/msh"
	"github.com/Konstantin8105/pow"
)

// Moment inertia calculation:
//
//	dA = ax*dy
//	Jxx = integral(y*y,dA)          // bending moments of inertia
//	Jyy = integral(x*x,dA)          // bending moments of inertia
//	Jxy = integral(x*y,dA)          // centrifugal moment of inertia
//	Jo  = integral(r^2,dA) = Jx+Jy  // polar moment of inertia
//	ko  = sqrt(Jo/A)
//
// Moment inertia for move system coordinate to (dx,dy):
//
//	Jxx` = Jxx + A*dy^2
//	Jyy` = Jyy + A*dx^2
//	Jxy` = Jxy + A*dx*dy
//
// Moment inertia for rotate system coordinate on angle O:
//
//	Ju  = (Jx+Jy)/2 + (Jx-Jy)/2*cos(2*O) - Jxy*sin(2*O)
//	Jv  = (Jx+Jy)/2 - (Jx-Jy)/2*cos(2*O) + Jxy*sin(2*O)
//	Juv = (Jx-Jy)/2*sin(2*O)+Jxy*cos(2*O)
//	Jo  = Ju+Jv = Jx+Jy
//
// Max/min moment inertia:
//
//	tan(2*O) = -2.0*Jxy/(Jx-Jy)
//	Jmax,min = (Jx+Jy)/2 +- sqrt(((Jx-Jy)/2)^2+Jxy^2)
//
// Matrix of moment inertia:
//
//	[ Jxx -Jxy]
//	[-Jxy  Jyy]
//
// First moment of area:
//
//	Sx = integral(y,dA)
//	Sy = integral(x,dA)
type BendingProperty struct {
	// https://en.wikipedia.org/wiki/First_moment_of_area
	// https://en.wikipedia.org/wiki/Second_moment_of_area
	// https://en.wikipedia.org/wiki/Section_modulus
	Jxx, Ymax, Wx, Rx, Sx, WxPlastic float64 // bending moments of inertia
	Jyy, Xmax, Wy, Ry, Sy, WyPlastic float64 // bending moments of inertia
	Jxy                              float64 // centrifugal moment of inertia
	Jo, Ro                           float64 // polar moment of inertia

	// TODO
	// Sx, Sy float64 // first moment of area
	// TODO : tau_xz = (Vy * Qz) / (Jz * t)
	// TODO : shear_area = (Jz * t) / Qz is maximal on each sections
	// Example http://www.learneasy.info/MDME/MEMmods/MEM09155A-CAE/050-Shear-in-Bending/shear-in-bending.html:
	// I-section
	//		height = 150
	//		width  = 100
	//		tw     = 10
	//		tf     = 10
	// 	J = 1.16e7 mm4
	// section A-A on 30 mm from neutral axe
	// if V = 25kN, then Q = 86625 mm3 and tau = 18.6MPa
	//
	// section B-B on 70 mm from neutral axe
	// in flange b = 100 mm, tau = 1.5 MPa
	// in web b = 10 mm, tau = 15 MPa
	// TODO: check on circle, ring, T-section
	// See https://engineering.stackexchange.com/questions/7989/shear-area-of-atypical-section
	// TODO https://www.ae.msstate.edu/tupas/SA2/Course.html
}

func (b BendingProperty) String() string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 3, ' ', tabwriter.TabIndent)
	// by axe X
	fmt.Fprintf(w, "By axe\tX\t.\n")
	fmt.Fprintf(w, "Jxx\t%s\tMoment inertia by axe X\n", efmt.Sprint(b.Jxx))
	fmt.Fprintf(w, "Ymax\t%s\tMaximal distance from axe X\n", efmt.Sprint(b.Ymax))
	fmt.Fprintf(w, "Wx\t%s\tElastic moment resistance around axe X\n", efmt.Sprint(b.Wx))
	fmt.Fprintf(w, "Rx\t%s\tRadius inertia around axe X\n", efmt.Sprint(b.Rx))
	fmt.Fprintf(w, "WxPlastic\t%s\tPlastic moment resistance around axe X\n", efmt.Sprint(b.WxPlastic))
	// by axe Y
	fmt.Fprintf(w, "By axe\tY\t.\n")
	fmt.Fprintf(w, "Jyy\t%s\tMoment inertia by axe Y\n", efmt.Sprint(b.Jyy))
	fmt.Fprintf(w, "Xmax\t%s\tMaximal distance from axe Y\n", efmt.Sprint(b.Xmax))
	fmt.Fprintf(w, "Wy\t%s\tElastic moment resistance around axe Y\n", efmt.Sprint(b.Wy))
	fmt.Fprintf(w, "Ry\t%s\tRadius inertia around axe Y\n", efmt.Sprint(b.Ry))
	fmt.Fprintf(w, "WyPlastic\t%s\tPlastic moment resistance around axe Y\n", efmt.Sprint(b.WyPlastic))
	// other
	fmt.Fprintf(w, "By axe\tOther\t.\n")
	fmt.Fprintf(w, "Jxy\t%s\tCentrifugal moment inertia by axe X-Y\n", efmt.Sprint(b.Jxy))
	// polar
	fmt.Fprintf(w, "By axe\tPolar\t.\n")
	fmt.Fprintf(w, "Jo\t%s\tPolar moment inertia\n", efmt.Sprint(b.Jo))
	fmt.Fprintf(w, "Ro\t%s\tPolar radius moment inertia\n", efmt.Sprint(b.Ro))

	fmt.Fprintf(w, "\n")
	w.Flush()
	return buf.String()
}

// In principal axes, that are rotated by an angle θ relative
// to original centroidal ones x,y, the product of inertia becomes zero.
func (b *BendingProperty) Alpha() float64 {
	angle := math.Atan(-2.0*b.Jxy/(b.Jxx-b.Jyy)) / 2.0
	angle += math.Pi / 2.0
	return angle
}

func (b *BendingProperty) Calculate(mesh msh.Msh) {
	A, _ := Area(mesh)
	calc := func() (j, h, w, r, wpl, s float64) {
		j = Jxx(mesh)
		h = Ymax(mesh)
		w = j / h
		r = math.Sqrt(j / A)
		wpl = WxPlastic(mesh)
		s = Sx(mesh)
		return
	}

	const perp float64 = math.Pi / 2.0 // 90 degree

	b.Jxx, b.Ymax, b.Wx, b.Rx, b.WxPlastic, b.Sx = calc()
	RotateXOY(&mesh, perp)
	b.Jyy, b.Xmax, b.Wy, b.Ry, b.WyPlastic, b.Sy = calc()
	RotateXOY(&mesh, -perp)
	b.Jxy = Jxy(mesh)
	b.Jo = b.Jxx + b.Jyy
	b.Ro = math.Sqrt(b.Jo / A)
}

type Property struct {
	Name string

	X, Y  float64 // location of center point
	Alpha float64 // angle from base coordinates

	A float64 // area

	// Property at base point
	AtBasePoint BendingProperty

	// Property at center of section
	AtCenterPoint BendingProperty

	// Property at center of section with rotation
	//	* minimal moment inertia on axe x
	//	* maximal moment inertia on axe y
	OnSectionAxe BendingProperty

	// TODO: torsion property
	// TODO: shear area
	// TODO: polar moment inertia
	// TODO: check on local buckling
	// TODO: center of shear
}

func (p Property) GetName() string {
	return p.Name
}

func (p Property) String() string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 3, ' ', tabwriter.TabIndent)
	if p.Name != "" {
		fmt.Fprintf(w, "Name: %s\n", p.Name)
	} else {
		fmt.Fprintf(w, "Name: %s\n", "Undefined")
	}
	fmt.Fprintf(w, "X\t%s\tlocation center of mass by axe X\n", efmt.Sprint(p.X))
	fmt.Fprintf(w, "Y\t%s\tlocation center of mass by axe Y\n", efmt.Sprint(p.Y))
	fmt.Fprintf(w, "Alpha\t%s\tAngle from base coordinates\n", efmt.Sprint(p.Alpha))
	fmt.Fprintf(w, "A\t%s\tArea of section\n", efmt.Sprint(p.A))
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "Bending property: At base point\n%s", p.AtBasePoint)
	fmt.Fprintf(w, "Bending property: At center point\n%s", p.AtCenterPoint)
	fmt.Fprintf(w, "Bending property: On section axe\n%s", p.OnSectionAxe)
	fmt.Fprintf(w, "\n")
	w.Flush()
	return buf.String()
}

const (
	Eps     = 1e-6
	IterMax = 200
)

// Area3node return area of triangle by coordinates
func Area3node(na, nb, nc msh.Node) float64 {
	// https://en.wikipedia.org/wiki/Triangle#Computing_the_area_of_a_triangle
	// Using Heron's formula
	var (
		a = math.Sqrt(pow.E2(na.Coord[0]-nb.Coord[0]) + pow.E2(na.Coord[1]-nb.Coord[1]) + pow.E2(na.Coord[2]-nb.Coord[2]))
		b = math.Sqrt(pow.E2(nc.Coord[0]-nb.Coord[0]) + pow.E2(nc.Coord[1]-nb.Coord[1]) + pow.E2(nc.Coord[2]-nb.Coord[2]))
		c = math.Sqrt(pow.E2(nc.Coord[0]-na.Coord[0]) + pow.E2(nc.Coord[1]-na.Coord[1]) + pow.E2(na.Coord[2]-nc.Coord[2]))
		s = (a + b + c) / 2.0 // semiperimeter
	)
	return math.Sqrt(s * (s - a) * (s - b) * (s - c))
}

// The centroid of a triangle is the point of intersection of its medians
func Center3node(na, nb, nc msh.Node) (center msh.Node) {
	// https://en.wikipedia.org/wiki/Centroid#Of_a_triangle
	center.Coord[0] = (na.Coord[0] + nb.Coord[0] + nc.Coord[0]) / 3.0
	center.Coord[1] = (na.Coord[1] + nb.Coord[1] + nc.Coord[1]) / 3.0
	center.Coord[2] = (na.Coord[2] + nb.Coord[2] + nc.Coord[2]) / 3.0
	return
}

func SortByY(na, nb, nc msh.Node) (_, _, _ msh.Node) {
	// sorting point by Y
	// Expect: na.Coord[1] <= nb.Coord[1] <= nc.Y
	if na.Coord[1] <= nb.Coord[1] {
		// do nothing
	} else {
		na, nb = nb, na // swap
	}
	// now : na.Coord[1] <= nb.Coord[1] and nc is ?
	switch {
	case nb.Coord[1] <= nc.Coord[1]:
		// do nothing
	case na.Coord[1] <= nc.Coord[1] && nc.Coord[1] <= nb.Coord[1]:
		nb, nc = nc, nb // swap
	case nc.Coord[1] <= na.Coord[1]:
		na, nb, nc = nc, na, nb // swap
	}

	return na, nb, nc
}

func Jx3node(na, nb, nc msh.Node) (j float64) {
	// calculate of moment of inertia from center point
	defer func() {
		c := Center3node(na, nb, nc)
		a := Area3node(na, nb, nc)
		j += a * pow.E2(c.Coord[1])
	}()

	switch {
	case na.Coord[1] == nb.Coord[1]:
		return J(na.Coord[0]-nb.Coord[0], nc.Coord[1]-na.Coord[1])
	case na.Coord[1] == nc.Coord[1]:
		return J(na.Coord[0]-nc.Coord[0], nb.Coord[1]-na.Coord[1])
	case nb.Coord[1] == nc.Coord[1]:
		return J(nb.Coord[0]-nc.Coord[0], na.Coord[1]-nb.Coord[1])
	}

	// sorting point by Y
	// na.Coord[1] < nb.Coord[1] < nc.Y
	na, nb, nc = SortByY(na, nb, nc)

	// find temp point X,Y between point `a` and `c`
	var temp = nb
	temp.Coord[0] = na.Coord[0] + (nc.Coord[0]-na.Coord[0])*(nb.Coord[1]-na.Coord[1])/(nc.Coord[1]-na.Coord[1])
	temp.Coord[1] = nb.Coord[1]

	// calculate moment of inertia for 2 triangles
	return J(nc.Coord[1]-nb.Coord[1], nb.Coord[0]-temp.Coord[0]) + J(nb.Coord[1]-na.Coord[1], nb.Coord[0]-temp.Coord[0])
}

// J return moment inertia of triangle:
//
//	b - size of triangle base
//	h - heigth of triangle
func J(b, h float64) float64 {
	return math.Abs(b * pow.E3(h) / 12.0)
}

func Calculate(g Geor) (p *Property, err error) {
	var mesh *msh.Msh
	p = new(Property)
	p.Name = g.GetName()

	{ // calculate area and choose precition
		lastArea := 0.0
		var prec float64 = 0.1 // TODO: auto finding
		var center msh.Node
		for iter := 0; iter < IterMax; iter++ {
			prec /= 2.0
			// choose precition by area
			mesh, err = msh.New(g.Geo(prec))
			if err != nil {
				return
			}
			if iter == 0 {
				lastArea, _ = Area(*mesh) // center is no need
				continue
			}
			if lastArea <= 0 {
				err = fmt.Errorf("Area is not valid: %e", lastArea)
				return
			}
			p.A, center = Area(*mesh)
			if math.Abs((p.A-lastArea)/lastArea) < Eps {
				break
			}
			lastArea = p.A
		}
		p.X = center.Coord[0]
		p.Y = center.Coord[1]
	}

	for i, point := range mesh.Nodes {
		if point.Coord[2] != 0 {
			err = fmt.Errorf("Coordinate Z of point %d is not zero: %f", i, point.Coord[2])
			return
		}
	}

	// calculate at the base point
	p.AtBasePoint.Calculate(*mesh)

	// calculate at the center point
	MoveXOY(mesh, -p.X, -p.Y)

	// calculate at the center point
	p.AtCenterPoint.Calculate(*mesh)

	// calculate at the center point with Jx minimal moment of inertia
	p.Alpha = p.AtCenterPoint.Alpha()

	// rotate
	RotateXOY(mesh, -p.Alpha)

	// calculate at the center point and rotate axes
	p.OnSectionAxe.Calculate(*mesh)

	return
}

func RotateXOY(m *msh.Msh, a float64) {
	for i := range m.Nodes {
		x, y := m.Nodes[i].Coord[0], m.Nodes[i].Coord[1]
		ampl := math.Sqrt(pow.E2(x) + pow.E2(y))
		angle := math.Atan2(y, x) + a
		m.Nodes[i].Coord[0] = ampl * math.Cos(angle)
		m.Nodes[i].Coord[1] = ampl * math.Sin(angle)
	}
}

func MoveXOY(m *msh.Msh, x, y float64) {
	for i := range m.Nodes {
		m.Nodes[i].Coord[0] += x
		m.Nodes[i].Coord[1] += y
	}
}

func Area(mesh msh.Msh) (Area float64, Center msh.Node) {
	for i := range mesh.Elements {
		if mesh.Elements[i].EType != msh.Triangle {
			continue
		}
		var (
			ns     = mesh.Elements[i].NodeId
			p0     = mesh.Nodes[mesh.GetNode(ns[0])]
			p1     = mesh.Nodes[mesh.GetNode(ns[1])]
			p2     = mesh.Nodes[mesh.GetNode(ns[2])]
			area   = Area3node(p0, p1, p2)
			center = Center3node(p0, p1, p2)
		)
		Center.Coord[0] = (area*center.Coord[0] + Area*Center.Coord[0]) / (Area + area)
		Center.Coord[1] = (area*center.Coord[1] + Area*Center.Coord[1]) / (Area + area)
		Area += area
	}

	return
}

func Ymax(mesh msh.Msh) float64 {
	var yMax float64
	for i := range mesh.Nodes {
		if i == 0 {
			yMax = math.Abs(mesh.Nodes[i].Coord[1])
		}
		if y := math.Abs(mesh.Nodes[i].Coord[1]); yMax < y {
			yMax = y
		}
	}
	return yMax
}

// first moment of area
// Sx = integral{y, dA)
func Sx(mesh msh.Msh) float64 {
	var S float64
	for i := range mesh.Elements {
		if mesh.Elements[i].EType != msh.Triangle {
			continue
		}
		var (
			ns = mesh.Elements[i].NodeId
			na = mesh.Nodes[mesh.GetNode(ns[0])]
			nb = mesh.Nodes[mesh.GetNode(ns[1])]
			nc = mesh.Nodes[mesh.GetNode(ns[2])]
			// center mass of triangle
			c = Center3node(na, nb, nc)
			// area of triangle
			a = Area3node(na, nb, nc)
		)
		S += math.Abs(c.Coord[1] * a)
	}
	return S
}

func Jxx(mesh msh.Msh) float64 {
	var J float64
	for i := range mesh.Elements {
		if mesh.Elements[i].EType != msh.Triangle {
			continue
		}
		var (
			ns = mesh.Elements[i].NodeId
			p0 = mesh.Nodes[mesh.GetNode(ns[0])]
			p1 = mesh.Nodes[mesh.GetNode(ns[1])]
			p2 = mesh.Nodes[mesh.GetNode(ns[2])]
		)
		J += Jx3node(p0, p1, p2)
	}
	if J < 0 {
		J = 0.0
	}
	return J
}

func Jxy(mesh msh.Msh) float64 {
	// See: https://en.wikipedia.org/wiki/Second_moment_of_area
	// Section: "Any polygon"
	var J float64
	for i := range mesh.Elements {
		if mesh.Elements[i].EType != msh.Triangle {
			continue
		}
		ns := mesh.Elements[i].NodeId
		p := [3]msh.Node{
			mesh.Nodes[mesh.GetNode(ns[0])],
			mesh.Nodes[mesh.GetNode(ns[1])],
			mesh.Nodes[mesh.GetNode(ns[2])],
		}
		if orientation(p[0], p[1], p[2]) == 2 {
			p[0], p[1] = p[1], p[0]
		}
		var ps [4]msh.Node
		for i := range p {
			ps[i] = p[i]
		}
		ps[3] = p[0]

		var jxy float64
		for i := 0; i < 3; i++ {
			jxy += (ps[i].Coord[0]*ps[i+1].Coord[1] - ps[i+1].Coord[0]*ps[i].Coord[1]) *
				(ps[i].Coord[0]*ps[i+1].Coord[1] + 2*ps[i].Coord[0]*ps[i].Coord[1] + 2*ps[i+1].Coord[0]*ps[i+1].Coord[1] + ps[i+1].Coord[0]*ps[i].Coord[1])
		}

		J += jxy
	}
	if J < 0 {
		J = 0.0
	}
	return 1.0 / 24.0 * J
}

// To find orientation of ordered triplet (p1, p2, p3).
// The function returns following values
// 0 --> p, q and r are colinear
// 1 --> Clockwise
// 2 --> Counterclockwise
func orientation(p1, p2, p3 msh.Node) int {
	// See 10th slides from following link for derivation
	// of the formula
	val := (p2.Coord[1]-p1.Coord[1])*(p3.Coord[0]-p2.Coord[0]) - (p2.Coord[0]-p1.Coord[0])*(p3.Coord[1]-p2.Coord[1])
	if val == 0 {
		return 0 // colinear
	}
	if val > 0 {
		return 1
	}
	return 2
}

func OnAxeX(a, b msh.Node) (c msh.Node) {
	c.Coord[0] = a.Coord[0] + (b.Coord[0]-a.Coord[0])*math.Abs(a.Coord[1]/(a.Coord[1]-b.Coord[1]))
	return
}

func WxPlastic(mesh msh.Msh) (w float64) {
	for i := range mesh.Elements {
		if mesh.Elements[i].EType != msh.Triangle {
			continue
		}
		var (
			ns = mesh.Elements[i].NodeId
			p  = [3]msh.Node{mesh.Nodes[mesh.GetNode(ns[0])],
				mesh.Nodes[mesh.GetNode(ns[1])],
				mesh.Nodes[mesh.GetNode(ns[2])],
			}
			sign [3]bool
		)
		p[0], p[1], p[2] = SortByY(p[0], p[1], p[2])
		for i := range p {
			sign[i] = math.Signbit(p[i].Coord[1])
		}
		var tr [][3]msh.Node
		switch {
		case sign[0] == sign[1] && sign[1] == sign[2]:
			tr = append(tr, p)

		case sign[0] != sign[1] && sign[1] == sign[2]:
			// find 2 point on axe X
			// between 0 and 1
			p01 := OnAxeX(p[0], p[1])
			// between 0 and 2
			p02 := OnAxeX(p[0], p[2])
			// triangles:
			tr = append(tr,
				[3]msh.Node{p[0], p01, p02},
				[3]msh.Node{p[1], p01, p02},
				[3]msh.Node{p[1], p02, p[2]},
			)

		case sign[0] == sign[1] && sign[1] != sign[2]:
			// find 2 point on axe X
			// between 2 and 0
			p20 := OnAxeX(p[2], p[0])
			// between 2 and 1
			p21 := OnAxeX(p[2], p[1])
			// triangles:
			tr = append(tr,
				[3]msh.Node{p[2], p20, p21},
				[3]msh.Node{p[1], p21, p20},
				[3]msh.Node{p[0], p20, p[1]},
			)
		}
		// calculate for one triangle
		for _, n := range tr {
			area := Area3node(n[0], n[1], n[2])
			center := Center3node(n[0], n[1], n[2])
			w += area * math.Abs(center.Coord[1])
		}
	}
	return
}
