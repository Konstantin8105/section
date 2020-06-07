package section_test

import (
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/Konstantin8105/msh"
	"github.com/Konstantin8105/section"
)

func TestAngle(t *testing.T){
	an := section.Angles[9]

	pr, err := section.Calculate(an)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(os.Stdout, "%#v\n", an)
	fmt.Fprintf(os.Stdout, "%#v\n", pr)
}


func ExampleUpn() {
	upn := section.UPNs[4]

	pr, err := section.Calculate(upn)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(os.Stdout, "%s\n", upn)
	fmt.Fprintf(os.Stdout, "%#v\n", pr)

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
