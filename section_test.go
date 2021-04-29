package section_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/Konstantin8105/msh"
	"github.com/Konstantin8105/section"
)

func ExampleGet() {
	g, err := section.Get("20B1-ASCM")
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stdout, "Type:%T\nDescription:%#v\n", g, g)
	// Output:
	// Type:section.Isection
	// Description:section.Isection{Name:"20B1-ASCM", H:0.2, B:0.1, Tw:0.0055, Tf:0.008, Radius:0.011}
}

func printJson(pr *section.Property) string {
	b, err := json.Marshal(*pr)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	err = json.Indent(&buf, b, " ", "\t")
	if err != nil {
		panic(err)
	}
	return buf.String()
}

// func Example() {
// 	upn := section.UPNs[4]
//
// 	pr, err := section.Calculate(upn)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Fprintf(os.Stdout, "%s\n",printJson(pr))
//
// 	// Output:
// }

func Test(t *testing.T) {
	tcs := []struct {
		na, nb, nc msh.Point
		area       float64
	}{
		{
			na:   msh.Point{X: 0, Y: 0},
			nb:   msh.Point{X: 1, Y: 0},
			nc:   msh.Point{X: 0, Y: 2},
			area: 1.0,
		},
		{
			na:   msh.Point{X: 0, Y: 0},
			nb:   msh.Point{X: -1, Y: 0},
			nc:   msh.Point{X: 0, Y: -2},
			area: 1.0,
		},
		{
			na:   msh.Point{X: 0, Y: 0},
			nb:   msh.Point{X: -1, Y: 0},
			nc:   msh.Point{X: 0, Y: 2},
			area: 1.0,
		},
		{
			na:   msh.Point{X: 0, Y: 0},
			nb:   msh.Point{X: 2, Y: 0},
			nc:   msh.Point{X: 0, Y: 4},
			area: 4.0,
		},
	}
	for i, tc := range tcs {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			a := section.Area3node(tc.na, tc.nb, tc.nc)
			if eps := 1e-6; math.Abs((a-tc.area)/a) > eps {
				t.Fatalf("%e != %e", a, tc.area)
			}
			t.Logf("Area = %e", a)
		})
	}
}

func TestSortByY(t *testing.T) {
	tcs := []struct {
		na, nb, nc msh.Point
	}{
		{na: msh.Point{Y: -1}, nb: msh.Point{Y: -1}, nc: msh.Point{Y: -1}},
		{na: msh.Point{Y: -1}, nb: msh.Point{Y: 0}, nc: msh.Point{Y: 1}},
		{na: msh.Point{Y: 1}, nb: msh.Point{Y: 0}, nc: msh.Point{Y: -1}},
		{na: msh.Point{Y: 1}, nb: msh.Point{Y: 2}, nc: msh.Point{Y: 3}},
		{na: msh.Point{Y: 3}, nb: msh.Point{Y: 2}, nc: msh.Point{Y: 1}},
		{na: msh.Point{Y: -1e-13}, nb: msh.Point{Y: -1}, nc: msh.Point{Y: -1}},
		{na: msh.Point{Y: -1e-13}, nb: msh.Point{Y: -1e-12}, nc: msh.Point{Y: -1e-111}},
		{na: msh.Point{Y: 0}, nb: msh.Point{Y: -1}, nc: msh.Point{Y: -1}},
	}
	for i, tc := range tcs {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			na, nb, nc := section.SortByY(tc.na, tc.nb, tc.nc)
			if !(na.Y <= nb.Y && nb.Y <= nc.Y) {
				t.Errorf("%v %v %v", na.Y, nb.Y, nc.Y)
			}
		})
	}
}

// TODO: add test for section: https://constructalia.arcelormittal.com/files/Sections_MB_ArcelorMittal_FR_EN_DE--7abaf280a4d4320516a471ede53f4adb.pdf

// Universal Beams (UB) sizes
// https://britishsteel.co.uk/media/40495/british-steel-sections-datasheets.pdf
func TestUB(t *testing.T) {
	table := `
//0 1  2  3  4  5    6     7     8   9   10    11   12    13   14   15    16   17
457 x 152 x 74 74.2 462.0 154.4 9.6 17.0 12.7 402.6 4.54 41.9 32891 1047 18.60 3.32
457 x 152 x 82 82.1 465.8 155.3 10.5 18.9 12.7 402.6 4.11 38.3 36806 1185 18.70 3.36
457 x 191 x 67 67.1 453.4 189.9 8.5 12.7 12.7 402.6 7.48 47.4 29597 1452 18.60 4.11
457 x 191 x 74 74.3 457.0 190.4 9.0 14.5 12.7 402.6 6.57 44.7 33536 1672 18.80 4.19
457 x 191 x 82 82.0 460.0 191.3 9.9 16.0 12.7 402.6 5.98 40.7 37268 1871 18.80 4.22
457 x 191 x 89 89.3 463.4 191.9 10.5 17.7 12.7 402.6 5.42 38.3 41232 2090 19.00 4.28

//18  19 20    21  22    23    24     25   26
1424 136 1637 213 0.873 29.7 0.5180 68.20 95.0 457 x 152 x 74
1580 153 1822 241 0.874 27.0 0.5920 92.00 105.0 457 x 152 x 82
1306 153 1481 238 0.873 37.2 0.7050 38.70 86.0 457 x 191 x 67
1468 176 1663 272 0.878 33.4 0.8180 53.60 95.1 457 x 191 x 74
1620 196 1842 304 0.878 30.5 0.9220 71.40 105.0 457 x 191 x 82
1780 218 2024 339 0.880 27.9 1.0400 93.30 114.0 457 x 191 x 89
`
	//
	lines := strings.Split(table, "\n")
	{
	again:
		for i := 0; i < len(lines); i++ {
			ts := strings.TrimSpace(lines[i])
			if len(ts) == 0 || ts[0] == '/' {
				lines = append(lines[:i], lines[i+1:]...)
				goto again
			}
		}
		size := len(lines) / 2
		for i := 0; i < size; i++ {
			lines[i] = lines[i] + " " + lines[i+size]
		}
		lines = lines[:size]
	}
	for _, l := range lines {
		fields := strings.Fields(l)
		name := strings.Join(fields[:5], "")
		t.Run(name, func(t *testing.T) {
			v := func(s string) float64 {
				f, err := strconv.ParseFloat(s, 64)
				if err != nil {
					t.Fatalf("%v", err)
				}
				return f
			}
			var is section.Isection
			is.Name = name
			is.H = v(fields[6]) * 1e-3
			is.B = v(fields[7]) * 1e-3
			is.Tf = v(fields[9]) * 1e-3
			is.Tw = v(fields[8]) * 1e-3
			is.Radius = v(fields[10]) * 1e-3

			pr, err := section.Calculate(is)
			if err != nil {
				t.Fatalf("%v", err)
				return
			}
			compare := func(pos int, f float64) {
				expect, err := strconv.ParseFloat(fields[pos], 64)
				if err != nil {
					t.Fatalf("On pos %d has error :%v", pos, err)
				}

				eps := 2.0 / 100.0 // 2%
				actEps := math.Abs((expect - f) / expect)
				if eps < actEps {
					t.Errorf("Not enougn precition for pos %2d: %8.2f != %8.2f. Prec = %5.2f %%",
						pos, f, expect, actEps*100)
				}
				eps = 0.5 / 100.0 // 0.5%
				if f < expect && eps < actEps {
					t.Errorf("Value is less on pos %d:  %8.2f <? %8.2f. Prec = %5.2f %%",
						pos, f, expect, actEps*100)
				}
			}
			compare(14, pr.AtCenterPoint.Jxx*1e8)
			compare(15, pr.AtCenterPoint.Jyy*1e8)

			compare(16, pr.AtCenterPoint.Rx*100)
			compare(17, pr.AtCenterPoint.Ry*100)

			compare(18, pr.AtCenterPoint.Wx*1e6)
			compare(19, pr.AtCenterPoint.Wy*1e6)

			compare(20, pr.AtCenterPoint.WxPlastic*1e6)
			compare(21, pr.AtCenterPoint.WyPlastic*1e6)

			compare(26, pr.A*1e4)
			// TODO: check torsion
		})
	}
}
