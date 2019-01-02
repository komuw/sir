package main

// DOCS: https://github.com/gonum/plot/blob/master/plotter/heat_test.go

import (
	"fmt"
	"log"
	"math"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/palette"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

/*
usage:
    go run examples/heatmap.go
*/

func main() {
	noOfAllRequests := 11
	lengthOfLargestRequest := 4
	Xdense := mat.NewDense(
		noOfAllRequests, lengthOfLargestRequest,
		[]float64{
			1, 2, 3, 4,
			1, 2, 3, 4,
			1, 2, 3, 4,
			1, 2, 3, 4,
			1, 2, 3, 4,
			1, 2, 3, 4,
			1, 2, 3, 4,
			5, 6, 7, 8,
			9, 10, 11, 12,
			13, 14, 15, 16,
			17, 18, 19, 20})

	m := offsetUnitGrid{
		XOffset: 0,
		YOffset: 0,
		Data:    Xdense}
	pal := palette.Heat(noOfAllRequests*lengthOfLargestRequest, 1)
	h := plotter.NewHeatMap(m, pal)

	p, err := plot.New()
	if err != nil {
		log.Fatal(err)
	}
	p.Title.Text = "Heat map"
	p.X.Tick.Marker = integerTicks{}
	p.Y.Tick.Marker = integerTicks{}

	p.Add(h)
	err = p.Save(6*vg.Inch, 4*vg.Inch, "examples/coolHeat.png")
	if err != nil {
		log.Fatal(err)
	}
}

type offsetUnitGrid struct {
	XOffset, YOffset float64
	Data             *mat.Dense
}

func (g offsetUnitGrid) Dims() (c, r int)   { r, c = g.Data.Dims(); return c, r }
func (g offsetUnitGrid) Z(c, r int) float64 { return g.Data.At(r, c) }
func (g offsetUnitGrid) X(c int) float64 {
	_, n := g.Data.Dims()
	if c < 0 || n <= c {
		panic("index out of range")
	}
	return float64(c) + g.XOffset
}
func (g offsetUnitGrid) Y(r int) float64 {
	m, _ := g.Data.Dims()
	if r < 0 || m <= r {
		panic("index out of range")
	}
	return float64(r) + g.YOffset
}

type integerTicks struct{}

func (integerTicks) Ticks(min, max float64) []plot.Tick {
	var t []plot.Tick
	for i := math.Trunc(min); i <= max; i++ {
		t = append(t, plot.Tick{Value: i, Label: fmt.Sprint(i)})
	}
	return t
}
