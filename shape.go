package section

import (
	"fmt"
	"math"

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
// Moment inertia for rotate system coordinate on angle O:
//	Ju  = (Jx+Jy)/2 + (Jx-Jy)/2*cos(2*O) - Jxy*sin(2*O)
//	Jv  = (Jx+Jy)/2 - (Jx-Jy)/2*cos(2*O) + Jxy*sin(2*O)
//	Juv = (Jx-Jy)/2*sin(2*O)+Jxy*cos(2*O)
//	Jo  = Ju+Jv = Jx+Jy
//
// Max/min moment inertia:
//	tan(2*O) = (-Jxy)/((Jx-Jy)/2)
//	Jmax,min = (Jx+Jy)/2 +- sqrt(((Jx-Jy)/2)^2+Jxy^2)
//
// Matrix of moment inertia:
//	[ Jxx -Jxy]
//	[-Jxy  Jyy]
//
type BendingProperty struct {
	Jxx, Ymax, Wx, Rx, WxPlastic float64
	Jyy, Xmax, Wy, Ry, WyPlastic float64
	Jxy                          float64 // TODO
	// TODO http://homes.civil.aau.dk/jc/FemteSemester/Beams3D.pdf
}

func (b *BendingProperty) Calculate(mesh msh.Msh) {
	A,_ := Area(mesh)
	calc := func() (j, h, w, r, wpl float64) {
		j, h = Jxx(mesh)
		w = j / h
		r = math.Sqrt(j / A)
		wpl = WxPlastic(mesh)
		return
	}

	const perp float64 = math.Pi/2.0 // 90 degree

	b.Jxx, b.Ymax, b.Wx, b.Rx, b.WxPlastic = calc()
	mesh.RotateXOY(perp)
	b.Jyy, b.Xmax, b.Wy, b.Ry, b.WyPlastic = calc()
	mesh.RotateXOY(-perp)
	b.Jxy = Jxy(mesh)
}

type Property struct {
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

const (
	Eps     = 1e-6
	IterMax = 200
)

// Area3node return area of triangle by coordinates
func Area3node(na, nb, nc msh.Point) float64 {
	// https://en.wikipedia.org/wiki/Triangle#Computing_the_area_of_a_triangle
	// Using Heron's formula
	var (
		a = math.Sqrt(pow.E2(na.X-nb.X) + pow.E2(na.Y-nb.Y) + pow.E2(na.Z-nb.Z))
		b = math.Sqrt(pow.E2(nc.X-nb.X) + pow.E2(nc.Y-nb.Y) + pow.E2(nc.Z-nb.Z))
		c = math.Sqrt(pow.E2(nc.X-na.X) + pow.E2(nc.Y-na.Y) + pow.E2(na.Z-nc.Z))
		s = (a + b + c) / 2.0 // semiperimeter
	)
	return math.Sqrt(s * (s - a) * (s - b) * (s - c))
}

// The centroid of a triangle is the point of intersection of its medians
func Center3node(na, nb, nc msh.Point) (center msh.Point) {
	// https://en.wikipedia.org/wiki/Centroid#Of_a_triangle
	center.X = (na.X + nb.X + nc.X) / 3.0
	center.Y = (na.Y + nb.Y + nc.Y) / 3.0
	center.Z = (na.Z + nb.Z + nc.Z) / 3.0
	return
}

func SortByY(na, nb, nc msh.Point) (_, _, _ msh.Point) {
	// sorting point by Y
	// Expect: na.Y <= nb.Y <= nc.Y
	if na.Y <= nb.Y {
		// do nothing
	} else {
		na, nb = nb, na // swap
	}
	// now : na.Y <= nb.Y and nc is ?
	switch {
	case nb.Y <= nc.Y:
		// do nothing
	case na.Y <= nc.Y && nc.Y <= nb.Y:
		nb, nc = nc, nb // swap
	case nc.Y <= na.Y:
		na, nb, nc = nc, na, nb // swap
	}

	return na, nb, nc
}

func Jx3node(na, nb, nc msh.Point) (j float64) {
	// calculate of moment of inertia from center point
	defer func() {
		c := Center3node(na, nb, nc)
		a := Area3node(na, nb, nc)
		j += a * pow.E2(c.Y)
	}()

	switch {
	case na.Y == nb.Y:
		return J(na.X-nb.X, nc.Y-na.Y)
	case na.Y == nc.Y:
		return J(na.X-nc.X, nb.Y-na.Y)
	case nb.Y == nc.Y:
		return J(nb.X-nc.X, na.Y-nb.Y)
	}

	// sorting point by Y
	// na.Y < nb.Y < nc.Y
	na, nb, nc = SortByY(na, nb, nc)

	// find temp point X,Y between point `a` and `c`
	var temp = nb
	temp.X = na.X + (nc.X-na.X)*(nb.Y-na.Y)/(nc.Y-na.Y)
	temp.Y = nb.Y

	// calculate moment of inertia for 2 triangles
	return J(nc.Y-nb.Y, nb.X-temp.X) + J(nb.Y-na.Y, nb.X-temp.X)
}

// J return moment inertia of triangle:
//	b - size of triangle base
//	h - heigth of triangle
func J(b, h float64) float64 {
	return math.Abs(b * pow.E3(h) / 12.0)
}

func Calculate(g interface{ Geo(prec float64) string }) (p *Property, err error) {
	var mesh *msh.Msh
	p = new(Property)

	{ // calculate area and choose precition
		lastArea := 0.0
		var prec float64 = 0.1 // TODO: auto finding
		var center msh.Point
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
		p.X = center.X
		p.Y = center.Y
	}

	for i, point := range mesh.Points {
		if point.Z != 0 {
			err = fmt.Errorf("Coordinate Z of point %d is not zero: %f", i, point.Z)
			return
		}
	}

	// calculate at the base point
	p.AtBasePoint.Calculate(*mesh)

	// calculate at the center point
	mesh.MoveXOY(-p.X, -p.Y)

	// calculate at the center point
	p.AtCenterPoint.Calculate(*mesh)

	// calculate at the center point with Jx minimal moment of inertia
	p.Alpha = func() (alpha float64) {
		points := make([]msh.Point, len(mesh.Points))
		copy(points, mesh.Points) // copy
		defer func() {
			copy(mesh.Points, points) // repair points
		}()
		// find minimimal J between 0 and Pi
		var (
			left  = 0.0
			right = math.Pi
			lastJ = math.MaxFloat64
		)
		// finding
		for da := math.Pi / 8.0; Eps <= da; da = da / 4.0 { // precision
			for a := left; a <= right; a += da {
				copy(mesh.Points, points)          // repair points
				mesh.RotateXOY(a)                  // rotate
				if J, _ := Jxx(*mesh); J < lastJ { // moment of inertia
					alpha, lastJ = a, J // store result
				}
			}
			left, right = alpha-da, alpha+da // new borders
		}
		return
	}()

	// rotate
	mesh.RotateXOY(-p.Alpha)

	// calculate at the center point and rotate axes
	p.OnSectionAxe.Calculate(*mesh)

	return
}

func Area(mesh msh.Msh) (Area float64, Center msh.Point) {
	for i := range mesh.Triangles {
		var (
			p      = mesh.PointsById(mesh.Triangles[i].PointsId)
			area   = Area3node(p[0], p[1], p[2])
			center = Center3node(p[0], p[1], p[2])
		)
		Center.X = (area*center.X + Area*Center.X) / (Area + area)
		Center.Y = (area*center.Y + Area*Center.Y) / (Area + area)
		Area += area
	}

	return
}

func Jxx(mesh msh.Msh) (j, yMax float64) {
	for i := range mesh.Triangles {
		p := mesh.PointsById(mesh.Triangles[i].PointsId)
		j += Jx3node(p[0], p[1], p[2])
	}
	for i := range mesh.Points {
		if i == 0 {
			yMax = math.Abs(mesh.Points[i].Y)
		}
		if y := math.Abs(mesh.Points[i].Y); yMax < y {
			yMax = y
		}
	}
	return
}

func Jxy(mesh msh.Msh) float64 {
	// TODO :
	// 	for i := range mesh.Triangles {
	// 		p := mesh.PointsById(mesh.Triangles[i].PointsId)
	// 		j += Jx3node(p[0], p[1], p[2])
	// 	}
	return -1
}

func OnAxeX(a, b msh.Point) (c msh.Point) {
	c.X = a.X + (b.X-a.X)*math.Abs(a.Y/(a.Y-b.Y))
	return
}

func WxPlastic(mesh msh.Msh) (w float64) {
	for i := range mesh.Triangles {
		var (
			p    = mesh.PointsById(mesh.Triangles[i].PointsId)
			sign [3]bool
		)
		p[0], p[1], p[2] = SortByY(p[0], p[1], p[2])
		for i := range p {
			sign[i] = math.Signbit(p[i].Y)
		}
		var tr [][3]msh.Point
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
				[3]msh.Point{p[0], p01, p02},
				[3]msh.Point{p[1], p01, p02},
				[3]msh.Point{p[1], p02, p[2]},
			)

		case sign[0] == sign[1] && sign[1] != sign[2]:
			// find 2 point on axe X
			// between 2 and 0
			p20 := OnAxeX(p[2], p[0])
			// between 2 and 1
			p21 := OnAxeX(p[2], p[1])
			// triangles:
			tr = append(tr,
				[3]msh.Point{p[2], p20, p21},
				[3]msh.Point{p[1], p21, p20},
				[3]msh.Point{p[0], p20, p[1]},
			)
		}
		// calculate for one triangle
		for _, n := range tr {
			area := Area3node(n[0], n[1], n[2])
			center := Center3node(n[0], n[1], n[2])
			w += area * math.Abs(center.Y)
		}
	}
	return
}
