package section

import (
	"fmt"
	"math"

	"github.com/Konstantin8105/msh"
	"github.com/Konstantin8105/pow"
)

// func area_4node(na,nb, nc, nd msh.Point ) float64 {
//     return 2*area_3node(na,nb,nc);
// 	//fabs(0.5*((nb.X - na.X)*(nc.Y - na.Y)- (nc.X - na.X)*(nb.Y - na.Y)))*2;
// };

type Property struct {
	A       float64 // area
	Elastic struct {
		AtBasePoint struct {
			Jx, Ymax, Wx float64
			Jy, Xmax, Wy float64
		}
		AtCenterPoint struct {
			X, Y         float64 // location of center point
			Jx, Ymax, Wx float64
			Jy, Xmax, Wy float64
		}
		OnSectionAxe struct {
			X, Y         float64 // location of center point
			Alpha        float64 // angle from base coordinates
			Jx, Ymax, Wx float64 // minimal moment inertia
			Jy, Xmax, Wy float64 // maximal moment inertia
		}
		Jt, Wt float64
	}
	Plastic struct {
		Wx, Wy float64
	}
	// TODO: shear area
	// TODO: radius inertia
	// TODO: polar moment inertia
	// TODO: check on local buckling
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
	if na.Y < nb.Y {
		// do nothing
	} else {
		na, nb = nb, na // swap
	}
	switch {
	case na.Y < nc.Y && nc.Y < nb.Y:
		nb, nc = nc, nb // swap
	case nc.Y < na.Y:
		na, nb, nc = nc, na, nb // swap
	}

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
	var (
		prec float64 = 0.01
		mesh *msh.Msh
	)
	p = new(Property)

	{ // calculate area and choose precition
		lastArea := 0.0
		var center msh.Point
		for iter := 0; iter < IterMax; iter++ {
			prec /= 2.0
			// choose precition by area
			mesh, err = msh.New(g.Geo(prec))
			if err != nil {
				return
			}
			if iter == 0 {
				lastArea, _ = p.Area(mesh)
				continue
			}
			if lastArea <= 0 {
				err = fmt.Errorf("Area is not valid: %e", lastArea)
				return
			}
			p.A, center = p.Area(mesh)
			if math.Abs((p.A-lastArea)/lastArea) < Eps {
				break
			}
			lastArea = p.A
		}
		p.Elastic.AtCenterPoint.X = center.X
		p.Elastic.AtCenterPoint.Y = center.Y
	}

	for i, point := range mesh.Points {
		if point.Z != 0 {
			err = fmt.Errorf("Coordinate Z of point %d is not zero: %f", i, point.Z)
			return
		}
	}

	calc := func() (j, h, w float64) {
		j, h = p.Jx(mesh)
		w = j / h
		// TODO plastic
		return
	}

	p.Elastic.AtBasePoint.Jx, p.Elastic.AtBasePoint.Ymax, p.Elastic.AtBasePoint.Wx = calc()
	mesh.RotateXOY90deg()
	p.Elastic.AtBasePoint.Jy, p.Elastic.AtBasePoint.Xmax, p.Elastic.AtBasePoint.Wy = calc()
	mesh.RotateXOY90deg()

	// TODO : find minimal Jx axe
	// TODO : rotate
	mesh.MoveXOY(-p.Elastic.AtCenterPoint.X, -p.Elastic.AtCenterPoint.Y)

	p.Elastic.AtCenterPoint.Jx, p.Elastic.AtCenterPoint.Ymax, p.Elastic.AtCenterPoint.Wx = calc()
	mesh.RotateXOY90deg()
	p.Elastic.AtCenterPoint.Jy, p.Elastic.AtCenterPoint.Xmax, p.Elastic.AtCenterPoint.Wy = calc()
	mesh.RotateXOY90deg()

	return
}

func (p Property) Area(mesh *msh.Msh) (Area float64, Center msh.Point) {
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

func (p Property) Jx(mesh *msh.Msh) (j, yMax float64) {
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

// func (s Property) X_Zero(n1, n2 msh.Point) {
// 	a := (n2.Y - n1.Y) / (n2.X - n1.X)
// 	b := n1.Y - a*n1.X
// 	return -b / a
// }
//
// func (s Property) Y_Zero(n1, n2 msh.Point) {
// 	a := (n2.Y - n1.Y) / (n2.X - n1.X)
// 	b := n1.Y - a*n1.X
// 	return b
// }
//
// func (s Property) Jx_node(n0, n1, n2 msh.Point) {
// 	var temp_n [3]msh.Point
// 	temp_n[0] = n0
// 	temp_n[1] = n1
// 	temp_n[2] = n2
// 	var (
// 		a     = area_3node(n0, n1, n2)
// 		X_MIN = temp_n[0].X
// 		Y_MIN = temp_n[0].Y
// 	)
// 	for i := 0; i < 3; i++ {
// 		if Y_MIN > temp_n[i].Y {
// 			Y_MIN = temp_n[i].Y
// 		}
// 		if X_MIN > temp_n[i].X {
// 			X_MIN = temp_n[i].X
// 		}
// 	}
//
// 	for i := 0; i < 3; i++ {
// 		temp_n[i].X -= X_MIN
// 		temp_n[i].Y -= Y_MIN
// 	}
//
// 	x_left := 0
// 	x_mid := 0
// 	x_right := 0
// 	if temp_n[0].X >= temp_n[1].X && temp_n[0].X > temp_n[2].X {
// 		x_right = 0
// 		if temp_n[1].X > temp_n[2].X {
// 			x_mid = 1
// 			x_left = 2
// 		} else {
// 			x_mid = 2
// 			x_left = 1
// 		}
// 	}
// 	if temp_n[1].X >= temp_n[0].X && temp_n[1].X > temp_n[2].X {
// 		x_right = 1
// 		if temp_n[0].X > temp_n[2].X {
// 			x_mid = 0
// 			x_left = 2
// 		} else {
// 			x_mid = 2
// 			x_left = 0
// 		}
// 	}
// 	if temp_n[2].X >= temp_n[1].X && temp_n[2].X > temp_n[0].X {
// 		x_right = 2
// 		if temp_n[0].X > temp_n[1].X {
// 			x_mid = 0
// 			x_left = 1
// 		} else {
// 			x_mid = 1
// 			x_left = 0
// 		}
// 	}
// 	if temp_n[x_left].X == temp_n[x_mid].X && temp_n[x_left].Y < temp_n[x_mid].Y {
// 		r := x_left
// 		x_left = x_mid
// 		x_mid = r
// 	}
// 	if temp_n[x_right].X == temp_n[x_mid].X && temp_n[x_right].Y < temp_n[x_mid].Y {
// 		r := x_right
// 		x_right = x_mid
// 		x_mid = r
// 	}
//
// 	typeL := 0
// 	y0 := temp_n[x_left].Y + (temp_n[x_right].Y-temp_n[x_left].Y)/
// 		(temp_n[x_right].X-temp_n[x_left].X)*(temp_n[x_mid].X-temp_n[x_left].X)
// 	if temp_n[x_mid].Y < y0 {
// 		typeL = 0
// 	} else {
// 		typeL = 1
// 	}
//
// 	var (
// 		jx            = -1e30
// 		Jx_left_mid   = Jx_node(temp_n[x_left], temp_n[x_mid])
// 		Jx_mid_right  = Jx_node(temp_n[x_mid], temp_n[x_right])
// 		Jx_left_right = Jx_node(temp_n[x_left], temp_n[x_right])
// 	)
//
// 	if typeL == 0 {
// 		jx = +Jx_left_right
// 		-Jx_left_mid
// 		-Jx_mid_right
// 	}
//
// 	if typeL == 1 {
// 		jx = -Jx_left_right
// 		+Jx_left_mid
// 		+Jx_mid_right
// 	}
//
// 	YC := (temp_n[0].Y + temp_n[1].Y + temp_n[2].Y) / 3.
// 	if jx < 1e-10 {
// 		jx = 0
// 	}
// 	if jx < 0 {
// 		print_name("jx is less NULL")
// 		printf("jx[%e]\n", jx)
// 	}
// 	if a < 0 {
// 		print_name("area is less NULL")
// 	}
// 	jx += -a*pow(YC, 2.) + a*pow(Y_MIN+YC, 2.) //*fabs(Y_MIN)/Y_MIN;
//
// 	return jx
// }
//
// func (s Property) Jx_node(n1, n2 msh.Point) {
//
// 	temp_n1 := n1
// 	temp_n2 := n2
// 	if n1.Y == 0 && n2.Y == 0 {
// 		return 0
// 	}
// 	if n1.X == n2.X {
// 		return 0
// 	}
// 	if n1.X == n2.X && n1.Y == n2.Y {
// 		print_name("STRANGE")
// 		WARNING()
// 		return 0
// 	}
//
// 	if temp_n1.X > temp_n2.X {
// 		swap(temp_n1.X, temp_n2.X)
// 		return Jx_node(temp_n1, temp_n2)
// 	}
// 	if temp_n1.Y > temp_n2.Y {
// 		swap(temp_n1.Y, temp_n2.Y)
// 		return Jx_node(temp_n1, temp_n2)
// 	}
//
// 	var (
// 		jx = 0.0
// 		a  = temp_n1.Y
// 		b  = fabs(temp_n2.X - temp_n1.X)
// 		h  = fabs(temp_n2.Y - temp_n1.Y)
// 	)
// 	if temp_n2.Y < a {
// 		print_name("WARNING: position a")
// 	}
// 	jx = (b*pow(a, 3.)/12. + a*b*pow(a/2., 2.)) + (b*pow(h, 3.)/12. + (b*h/2.)*pow(a, 2.))
// 	return fabs(jx)
// }
//
// func (s Property) CalcJ(angle) {
// 	J := 0.0
// 	mesh.RotatePointXOY(0, 0, Angle)
// 	for i := 0; i < mesh.elements.GetSize(); i++ {
// 		el := mesh.elements.Get(i)
// 		if el.ElmType == ELEMENT_TYPE_TRIANGLE {
// 			var p [3]msh.Point
// 			p[0] = mesh.nodes.Get(el.node[0] - 1)
// 			p[1] = mesh.nodes.Get(el.node[1] - 1)
// 			p[2] = mesh.nodes.Get(el.node[2] - 1)
// 			J += Jx_node(p[0], p[1], p[2])
// 		}
// 	}
// 	mesh.RotatePointXOY(0, 0, -Angle)
// 	return J
// }
//
// func (s Property) AngleWithMinimumJ(float64 step0, float64 _angle) {
// 	var (
// 		x0  = _angle - step0*1
// 		x1  = _angle + step0*0
// 		x2  = _angle + step0*1
// 		y0  = CalcJ(x0)
// 		y1  = CalcJ(x1)
// 		y2  = CalcJ(x2)
// 		eps = 1e-6
// 	)
// 	if GRADIANS(max(x0, x1, x2)-min(x0, x1, x2)) <= eps || max(y0, y1, y2)-min(y0, y1, y2) <= eps*min(y0, y1, y2) {
// 		if y0 == min(y0, y1, y2) {
// 			return x0
// 		} else if y1 == min(y0, y1, y2) {
// 			return x1
// 		} else {
// 			return x2
// 		}
// 	}
// 	if min(y0, y1, y2) == y0 {
// 		return AngleWithMinimumJ(step0, x0)
// 	} else if min(y0, y1, y2) == y2 {
// 		return AngleWithMinimumJ(step0, x2)
// 	} else {
// 		return AngleWithMinimumJ(step0/1.5, x1)
// 	}
// }
//
// func (s Property) CalculateOLD() {
// 	i := 0
// 	numEl = mesh.elements.GetSize()
// 	Xc := 0
// 	Yc := 0
// 	Area = 0
// 	for i := 0; i < numEl; i++ {
// 		el := mesh.elements.Get(i)
// 		if el.ElmType == ELEMENT_TYPE_TRIANGLE {
// 			var p [3]msh.Point
// 			p[0] = mesh.nodes.Get(el.node[0] - 1)
// 			p[1] = mesh.nodes.Get(el.node[1] - 1)
// 			p[2] = mesh.nodes.Get(el.node[2] - 1)
// 			xc := (p[0].X + p[1].X + p[2].X) / 3.
// 			yc := (p[0].Y + p[1].Y + p[2].Y) / 3.
// 			a := area_3node(p[0], p[1], p[2])
// 			Xc = (a*xc + Area*Xc) / (Area + a)
// 			Yc = (a*yc + Area*Yc) / (Area + a)
// 			Area += a
// 		}
// 	}
// 	numPoint = mesh.nodes.GetSize()
// 	for i = 0; i < numPoint; i++ {
// 		mesh.nodes.Get(i).X -= Xc
// 		mesh.nodes.Get(i).Y -= Yc
// 	}
//
// 	{
// 		Ymax := 0
// 		angle := 0
// 		mesh.RotatePointXOY(0, 0, +RADIANS(angle))
// 		for i = 0; i < mesh.nodes.GetSize(); i++ {
// 			Ymax = max(Ymax, fabs(mesh.nodes.Get(i).Y))
// 		}
// 		Jx_MomentInertia = CalcJ(0)
// 		Wx_MomentInertia = Jx_MomentInertia / Ymax
// 		mesh.RotatePointXOY(0, 0, -RADIANS(angle))
// 	}
// 	{
// 		Ymax := 0
// 		angle := 90
// 		mesh.RotatePointXOY(0, 0, +RADIANS(angle))
// 		for i = 0; i < mesh.nodes.GetSize(); i++ {
// 			Ymax = max(Ymax, fabs(mesh.nodes.Get(i).Y))
// 		}
// 		Jy_MomentInertia = CalcJ(0)
// 		Wy_MomentInertia = Jy_MomentInertia / Ymax
// 		mesh.RotatePointXOY(0, 0, -RADIANS(angle))
// 	}
// 	angleMinJ = GRADIANS(AngleWithMinimumJ(RADIANS(45.), 0.))
// 	{
// 		Ymax = 0
// 		angle = AngleMinJ
// 		mesh.RotatePointXOY(0, 0, +RADIANS(angle))
// 		for i = 0; i < mesh.nodes.GetSize(); i++ {
// 			Ymax = max(Ymax, fabs(mesh.nodes.Get(i).Y))
// 		}
// 		Jv_MomentInertia = CalcJ(0)
// 		Wv_MomentInertia = Jv_MomentInertia / Ymax
// 		mesh.RotatePointXOY(0, 0, -RADIANS(angle))
// 	}
// 	{
// 		Ymax = 0
// 		angle = AngleMinJ + 90
// 		mesh.RotatePointXOY(0, 0, +RADIANS(angle))
// 		for i = 0; i < mesh.nodes.GetSize(); i++ {
// 			Ymax = max(Ymax, fabs(mesh.nodes.Get(i).Y))
// 		}
// 		Ju_MomentInertia = CalcJ(0)
// 		Wu_MomentInertia = Ju_MomentInertia / Ymax
// 		mesh.RotatePointXOY(0, 0, -RADIANS(angle))
// 	}
// 	// WX_PLASTIC //
// 	{
// 		Wx_Plastic = 0
// 		angle = 0
// 		mesh.RotatePointXOY(0, 0, +RADIANS(angle))
// 		for i = 0; i < numEl; i++ {
// 			el = mesh.elements.Get(i)
// 			if el.ElmType == ELEMENT_TYPE_TRIANGLE {
// 				var p [3]msh.Point
// 				p[0] = mesh.nodes.Get(el.node[0] - 1)
// 				p[1] = mesh.nodes.Get(el.node[1] - 1)
// 				p[2] = mesh.nodes.Get(el.node[2] - 1)
// 				simple_case := false
// 				if (p[0].Y != 0 || p[1].Y != 0 || p[2].Y != 0) &&
// 					(p[0].Y/p[1].Y > 0 && p[0].Y/p[2].Y > 0 && p[1].Y/p[2].Y > 0) {
// 					Wx_Plastic += (fabs(p[0].Y+p[1].Y+p[2].Y) / 3.) * (area_3node(p[0], p[1], p[2]))
// 					simple_case = true
// 				}
// 				if !simple_case {
// 					if (p[0].Y == 0 || p[1].Y == 0) ||
// 						(p[1].Y == 0 || p[2].Y == 0) ||
// 						(p[2].Y == 0 || p[0].Y == 0) {
// 						Wx_Plastic += (fabs(p[0].Y+p[1].Y+p[2].Y) / 3.) * (area_3node(p[0], p[1], p[2]))
// 					} else if p[0].Y == 0 && p[1].Y/p[2].Y < 0 {
// 						var tmp msh.Point // intersect with zero line
// 						tmp.X = p[1].X + (p[2].X-p[1].X)*fabs(p[1].Y)/(fabs(p[2].Y)+fabs(p[1].Y))
// 						tmp.Y = 0
// 						tmp.z = 0.
// 						Wx_Plastic += (fabs(p[0].Y+p[1].Y+tmp.Y) / 3.) * (area_3node(p[0], p[1], tmp))
// 						Wx_Plastic += (fabs(p[0].Y+p[2].Y+tmp.Y) / 3.) * (area_3node(p[0], p[2], tmp))
// 					} else if p[1].Y == 0 && p[2].Y/p[0].Y < 0 {
// 						var tmp msh.Point // intersect with zero line
// 						tmp.X = p[0].X + (p[2].X-p[0].X)*fabs(p[0].Y)/(fabs(p[2].Y)+fabs(p[0].Y))
// 						tmp.Y = 0
// 						tmp.z = 0.
// 						Wx_Plastic += (fabs(p[1].Y+p[0].Y+tmp.Y) / 3.) * (area_3node(p[1], p[0], tmp))
// 						Wx_Plastic += (fabs(p[1].Y+p[2].Y+tmp.Y) / 3.) * (area_3node(p[1], p[2], tmp))
// 					} else if p[2].Y == 0 && p[1].Y/p[0].Y < 0 {
// 						var tmp msh.Point // intersect with zero line
// 						tmp.X = p[0].X + (p[1].X-p[0].X)*fabs(p[0].Y)/(fabs(p[1].Y)+fabs(p[0].Y))
// 						tmp.Y = 0
// 						tmp.z = 0.
// 						Wx_Plastic += (fabs(p[2].Y+p[0].Y+tmp.Y) / 3.) * (area_3node(p[2], p[0], tmp))
// 						Wx_Plastic += (fabs(p[2].Y+p[1].Y+tmp.Y) / 3.) * (area_3node(p[2], p[1], tmp))
// 					} else if p[0].Y/p[1].Y < 0 && p[0].Y/p[2].Y < 0 {
// 						var tmp msh.Point1
// 						tmp1.Y = 0
// 						tmp1.z = 0.
// 						tmp1.X = p[0].X + (p[1].X-p[0].X)*fabs(p[0].Y)/(fabs(p[1].Y)+fabs(p[0].Y))
// 						var tmp msh.Point2
// 						tmp2.Y = 0
// 						tmp2.z = 0.
// 						tmp2.X = p[0].X + (p[2].X-p[0].X)*fabs(p[0].Y)/(fabs(p[2].Y)+fabs(p[0].Y))
// 						Wx_Plastic += (fabs(p[0].Y+tmp1.Y+tmp2.Y) / 3.) * (area_3node(p[0], tmp1, tmp2))
// 						Wx_Plastic += (fabs(p[1].Y+tmp1.Y+tmp2.Y) / 3.) * (area_3node(p[1], tmp1, tmp2))
// 						Wx_Plastic += (fabs(p[2].Y+p[1].Y+tmp2.Y) / 3.) * (area_3node(p[2], p[1], tmp2))
// 					} else if p[1].Y/p[0].Y < 0 && p[1].Y/p[2].Y < 0 {
// 						var tmp msh.Point1
// 						tmp1.Y = 0
// 						tmp1.z = 0.
// 						tmp1.X = p[0].X + (p[1].X-p[0].X)*fabs(p[0].Y)/(fabs(p[1].Y)+fabs(p[0].Y))
// 						var tmp msh.Point2
// 						tmp2.Y = 0
// 						tmp2.z = 0.
// 						tmp2.X = p[1].X + (p[2].X-p[1].X)*fabs(p[1].Y)/(fabs(p[2].Y)+fabs(p[1].Y))
// 						Wx_Plastic += (fabs(p[1].Y+tmp1.Y+tmp2.Y) / 3.) * (area_3node(p[1], tmp1, tmp2))
// 						Wx_Plastic += (fabs(p[0].Y+tmp1.Y+tmp2.Y) / 3.) * (area_3node(p[0], tmp1, tmp2))
// 						Wx_Plastic += (fabs(p[2].Y+p[0].Y+tmp2.Y) / 3.) * (area_3node(p[2], p[0], tmp2))
// 					} else if p[2].Y/p[0].Y < 0 && p[2].Y/p[1].Y < 0 {
// 						var tmp msh.Point1
// 						tmp1.Y = 0
// 						tmp1.z = 0.
// 						tmp1.X = p[0].X + (p[2].X-p[0].X)*fabs(p[0].Y)/(fabs(p[2].Y)+fabs(p[0].Y))
// 						var tmp msh.Point2
// 						tmp2.Y = 0
// 						tmp2.z = 0.
// 						tmp2.X = p[1].X + (p[2].X-p[1].X)*fabs(p[1].Y)/(fabs(p[2].Y)+fabs(p[1].Y))
// 						Wx_Plastic += (fabs(p[2].Y+tmp1.Y+tmp2.Y) / 3.) * (area_3node(p[2], tmp1, tmp2))
// 						Wx_Plastic += (fabs(p[0].Y+tmp1.Y+tmp2.Y) / 3.) * (area_3node(p[0], tmp1, tmp2))
// 						Wx_Plastic += (fabs(p[1].Y+p[0].Y+tmp2.Y) / 3.) * (area_3node(p[1], p[0], tmp2))
// 					} else {
// 						printf("IO-")
// 					}
// 				}
// 			}
// 		}
// 		mesh.RotatePointXOY(0, 0, -RADIANS(angle))
// 	}
// 	{
// 		Wy_Plastic = 0
// 		angle = 90
// 		mesh.RotatePointXOY(0, 0, +RADIANS(angle))
// 		for i = 0; i < numEl; i++ {
// 			el = mesh.elements.Get(i)
// 			if el.ElmType == ELEMENT_TYPE_TRIANGLE {
// 				var p [3]msh.Point
// 				p[0] = mesh.nodes.Get(el.node[0] - 1)
// 				p[1] = mesh.nodes.Get(el.node[1] - 1)
// 				p[2] = mesh.nodes.Get(el.node[2] - 1)
// 				simple_case := false
// 				if (p[0].Y != 0 || p[1].Y != 0 || p[2].Y != 0) &&
// 					(p[0].Y/p[1].Y > 0 && p[0].Y/p[2].Y > 0 && p[1].Y/p[2].Y > 0) {
// 					Wy_Plastic += (fabs(p[0].Y+p[1].Y+p[2].Y) / 3.) * (area_3node(p[0], p[1], p[2]))
// 					simple_case = true
// 				}
// 				if !simple_case {
// 					if (p[0].Y == 0 || p[1].Y == 0) ||
// 						(p[1].Y == 0 || p[2].Y == 0) ||
// 						(p[2].Y == 0 || p[0].Y == 0) {
// 						Wy_Plastic += (fabs(p[0].Y+p[1].Y+p[2].Y) / 3.) * (area_3node(p[0], p[1], p[2]))
// 					} else if p[0].Y == 0 && p[1].Y/p[2].Y < 0 {
// 						var tmp msh.Point // intersect with zero line
// 						tmp.X = p[1].X + (p[2].X-p[1].X)*fabs(p[1].Y)/(fabs(p[2].Y)+fabs(p[1].Y))
// 						tmp.Y = 0
// 						tmp.z = 0.
// 						Wy_Plastic += (fabs(p[0].Y+p[1].Y+tmp.Y) / 3.) * (area_3node(p[0], p[1], tmp))
// 						Wy_Plastic += (fabs(p[0].Y+p[2].Y+tmp.Y) / 3.) * (area_3node(p[0], p[2], tmp))
// 					} else if p[1].Y == 0 && p[2].Y/p[0].Y < 0 {
// 						var tmp msh.Point // intersect with zero line
// 						tmp.X = p[0].X + (p[2].X-p[0].X)*fabs(p[0].Y)/(fabs(p[2].Y)+fabs(p[0].Y))
// 						tmp.Y = 0
// 						tmp.z = 0.
// 						Wy_Plastic += (fabs(p[1].Y+p[0].Y+tmp.Y) / 3.) * (area_3node(p[1], p[0], tmp))
// 						Wy_Plastic += (fabs(p[1].Y+p[2].Y+tmp.Y) / 3.) * (area_3node(p[1], p[2], tmp))
// 					} else if p[2].Y == 0 && p[1].Y/p[0].Y < 0 {
// 						var tmp msh.Point // intersect with zero line
// 						tmp.X = p[0].X + (p[1].X-p[0].X)*fabs(p[0].Y)/(fabs(p[1].Y)+fabs(p[0].Y))
// 						tmp.Y = 0
// 						tmp.z = 0.
// 						Wy_Plastic += (fabs(p[2].Y+p[0].Y+tmp.Y) / 3.) * (area_3node(p[2], p[0], tmp))
// 						Wy_Plastic += (fabs(p[2].Y+p[1].Y+tmp.Y) / 3.) * (area_3node(p[2], p[1], tmp))
// 					} else if p[0].Y/p[1].Y < 0 && p[0].Y/p[2].Y < 0 {
// 						var tmp msh.Point1
// 						tmp1.Y = 0
// 						tmp1.z = 0.
// 						tmp1.X = p[0].X + (p[1].X-p[0].X)*fabs(p[0].Y)/(fabs(p[1].Y)+fabs(p[0].Y))
// 						var tmp msh.Point2
// 						tmp2.Y = 0
// 						tmp2.z = 0.
// 						tmp2.X = p[0].X + (p[2].X-p[0].X)*fabs(p[0].Y)/(fabs(p[2].Y)+fabs(p[0].Y))
// 						Wy_Plastic += (fabs(p[0].Y+tmp1.Y+tmp2.Y) / 3.) * (area_3node(p[0], tmp1, tmp2))
// 						Wy_Plastic += (fabs(p[1].Y+tmp1.Y+tmp2.Y) / 3.) * (area_3node(p[1], tmp1, tmp2))
// 						Wy_Plastic += (fabs(p[2].Y+p[1].Y+tmp2.Y) / 3.) * (area_3node(p[2], p[1], tmp2))
// 					} else if p[1].Y/p[0].Y < 0 && p[1].Y/p[2].Y < 0 {
// 						var tmp msh.Point1
// 						tmp1.Y = 0
// 						tmp1.z = 0.
// 						tmp1.X = p[0].X + (p[1].X-p[0].X)*fabs(p[0].Y)/(fabs(p[1].Y)+fabs(p[0].Y))
// 						var tmp msh.Point2
// 						tmp2.Y = 0
// 						tmp2.z = 0.
// 						tmp2.X = p[1].X + (p[2].X-p[1].X)*fabs(p[1].Y)/(fabs(p[2].Y)+fabs(p[1].Y))
// 						Wy_Plastic += (fabs(p[1].Y+tmp1.Y+tmp2.Y) / 3.) * (area_3node(p[1], tmp1, tmp2))
// 						Wy_Plastic += (fabs(p[0].Y+tmp1.Y+tmp2.Y) / 3.) * (area_3node(p[0], tmp1, tmp2))
// 						Wy_Plastic += (fabs(p[2].Y+p[0].Y+tmp2.Y) / 3.) * (area_3node(p[2], p[0], tmp2))
// 					} else if p[2].Y/p[0].Y < 0 && p[2].Y/p[1].Y < 0 {
// 						var tmp msh.Point1
// 						tmp1.Y = 0
// 						tmp1.z = 0.
// 						tmp1.X = p[0].X + (p[2].X-p[0].X)*fabs(p[0].Y)/(fabs(p[2].Y)+fabs(p[0].Y))
// 						var tmp msh.Point2
// 						tmp2.Y = 0
// 						tmp2.z = 0.
// 						tmp2.X = p[1].X + (p[2].X-p[1].X)*fabs(p[1].Y)/(fabs(p[2].Y)+fabs(p[1].Y))
// 						Wy_Plastic += (fabs(p[2].Y+tmp1.Y+tmp2.Y) / 3.) * (area_3node(p[2], tmp1, tmp2))
// 						Wy_Plastic += (fabs(p[0].Y+tmp1.Y+tmp2.Y) / 3.) * (area_3node(p[0], tmp1, tmp2))
// 						Wy_Plastic += (fabs(p[1].Y+p[0].Y+tmp2.Y) / 3.) * (area_3node(p[1], p[0], tmp2))
// 					} else {
// 						printf("IO-")
// 					}
// 				}
// 			}
// 		}
// 		mesh.RotatePointXOY(0, 0, -RADIANS(angle))
// 	}
// }
