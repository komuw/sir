package sir

import (
	"fmt"
	"image/color"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"

	"gonum.org/v1/plot/vg/draw"

	"github.com/pkg/errors"
	"time"
)

func PlotPlainScatter(labelsmap map[int]int, noOfAllRequests int, labels []int, nclusters int, X *mat.Dense, appendName string) error {
	now := time.Now()
	// Save the plot to a PNG file.
	pngfile := appendName + "_" + now.Format("Jan_2_2006_15_04_05") + ".png"

	// plot result
	pt, err := plot.New()
	if err != nil {
		return errors.Wrap(err, "error instantiating plot")

	}
	pt.Add(plotter.NewGrid())
	pt.Title.Text = fmt.Sprintf("Estimated number of clusters: %d", nclusters)

	for cl := range labelsmap {
		var data plotter.XYs
		for sample := 0; sample < noOfAllRequests; sample++ {
			if labels[sample] == cl {
				data = append(data, struct{ X, Y float64 }{X.At(sample, 0), X.At(sample, 1)})
			}
		}
		s, err := plotter.NewScatter(data)
		if err != nil {
			return errors.Wrap(err, "error instantiating plotter.NewScatter")
		}
		var color0 color.RGBA
		switch cl {
		case -1:
			color0 = color.RGBA{0, 0, 0, 255}
		case 0:
			color0 = color.RGBA{176, 0, 0, 255}
		case 1:
			color0 = color.RGBA{0, 176, 0, 255}
		case 2:
			color0 = color.RGBA{0, 0, 176, 255}
		}
		s.GlyphStyle.Color = color0
		s.GlyphStyle.Shape = draw.CircleGlyph{}
		pt.Add(s)
		// p.Legend.Add(log.Sprintf("scatter %d", cl), s)

	}

	err = pt.Save(6*vg.Inch, 4*vg.Inch, pngfile)
	if err != nil {
		return errors.Wrap(err, "error saving png")
	}

	return nil
}
