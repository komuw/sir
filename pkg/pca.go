package sir

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/pkg/errors"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

func PlotResultsPCA(noOfRequests int, X *mat.Dense, nclusters int, appendName string) error {
	now := time.Now()
	// Save the plot to a PNG file.
	pngfile := appendName + "_" + now.Format("Jan_02_2006_15_04_05") + ".png"

	// plot result
	pt, err := plot.New()
	if err != nil {
		return errors.Wrap(err, "error instantiating plot")

	}
	pt.Add(plotter.NewGrid())
	pt.Title.Text = fmt.Sprintf("%s: number of clusters: %d", appendName, nclusters)

	// type XYs []struct{ X, Y float64 }
	var data plotter.XYs
	for sample := 0; sample < noOfRequests; sample++ {
		data = append(data, struct{ X, Y float64 }{X.At(sample, 0), X.At(sample, 1)})
	}

	s, err := plotter.NewScatter(data)
	if err != nil {
		return errors.Wrap(err, "error instantiating plotter.NewScatter")
	}
	color0 := color.RGBA{176, 0, 0, 255}
	s.GlyphStyle.Color = color0
	s.GlyphStyle.Shape = draw.CircleGlyph{}
	pt.Add(s)
	// p.Legend.Add(log.Sprintf("scatter %d", cl), s)

	log.Println("start save png")
	// TODO: save is hanging. fix it
	err = pt.Save(6*vg.Inch, 4*vg.Inch, pngfile)
	if err != nil {
		return errors.Wrap(err, "error saving png")
	}
	log.Println("end save png")

	return nil
}

func FindPCA(X *mat.Dense, d int) *mat.Dense {
	// Calculate the principal component direction vectors and variances.
	var pc stat.PC
	ok := pc.PrincipalComponents(X, nil)
	if !ok {
		log.Fatal(errors.New("unable to get PrincipalComponents"))
	}
	// log.Println("variances ", pc.VarsTo(nil))

	k := 2
	var proj mat.Dense
	proj.Mul(X, pc.VectorsTo(nil).Slice(0, d, 0, k))

	return &proj
}
