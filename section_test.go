package section_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/Konstantin8105/msh"
	"github.com/Konstantin8105/section"
)

func TestAngle(t *testing.T) {
	an := section.Angles[9]

	pr, err := section.Calculate(an)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(os.Stdout, "%#v\n", an)
	fmt.Fprintf(os.Stdout, printJson(pr))
}

func ExamplePlate() {
	pl := section.Rectangle{
		H:   0.100,
		Thk: 0.010,
	}

	pr, err := section.Calculate(pl)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stdout, printJson(pr))

	// Output:
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

func ExampleUpn() {
	upn := section.UPNs[4]

	pr, err := section.Calculate(upn)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stdout, printJson(pr))

	// Output:
}

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
		{ na: msh.Point{Y: -1}, nb: msh.Point{Y: -1}, nc: msh.Point{Y: -1}, },
		{ na: msh.Point{Y: -1}, nb: msh.Point{Y: 0}, nc: msh.Point{Y: 1}, },
		{ na: msh.Point{Y: 1}, nb: msh.Point{Y: 0}, nc: msh.Point{Y: -1}, },
		{ na: msh.Point{Y: 1}, nb: msh.Point{Y: 2}, nc: msh.Point{Y: 3}, },
		{ na: msh.Point{Y: 3}, nb: msh.Point{Y: 2}, nc: msh.Point{Y: 1}, },
		{ na: msh.Point{Y: -1e-13}, nb: msh.Point{Y: -1}, nc: msh.Point{Y: -1}, },
		{ na: msh.Point{Y: -1e-13}, nb: msh.Point{Y: -1e-12}, nc: msh.Point{Y: -1e-111}, },
		{ na: msh.Point{Y: 0}, nb: msh.Point{Y: -1}, nc: msh.Point{Y: -1}, },
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
