package heart

import (
	"fmt"
	"image/color"
	"log"

	"github.com/pa-m/sklearn/cluster"
	"github.com/pa-m/sklearn/datasets"
	"github.com/pa-m/sklearn/preprocessing"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"

	"gonum.org/v1/plot/vg/draw"

	"github.com/pkg/errors"
	"time"
)

/*
usage:
  go run -race go_dbscan/go_dbscan.go
*/

// TODO: we should be able to handle requests that are not of equal size.
// Our current dbscan code can example only support something like;
// len("test out the server")  == len("something different")

func findClusterMembers(labels []int, X *mat.Dense) error {
	/*
		for
			  X := mat.NewDense(noOfAllRequests, 2, []float64{1, 2, 33, 3, 644, 7, 5555, 5})
			  X.At(i, j) // At returns the element at row i, column j.
			  X.At(0, 0) == 1
			  X.At(0, 1) == 2
			  X.At(0, 2) == Does NOT exist
			  X.At(1, 0) == 33
			  X.At(3, 1) == 5
			  X.At(3, 2) == Does not exist
			  X.At(4, 0) == Does not exist
	*/

	// cluster_members :=  map[string][]int

	rows, columns := X.Caps()
	log.Println("rows, columns", rows, columns)
	for k := range labels {
		for i := 0; i < columns; i++ {
			zAt := X.At(k, i)
			_ = zAt
			// log.Println("zAt", zAt)
		}

	}
	return nil
}

func Run(noOfAllRequests int, lengthOfEachRequest int, allRequests []float64, Eps float64, MinSamples float64, autoGenerateSampleData bool, appendName string) {
	// adapted from http://scikit-learn.org/stable/_downloads/plot_dbscan.ipynb
	if lengthOfEachRequest <= 1 {
		err := errors.New("we cant create a matrix with no dimensions, ie X.At(x, y) will fail")
		log.Fatalf("\n%+v", err)
	}

	X := getX(noOfAllRequests, lengthOfEachRequest, allRequests)
	if autoGenerateSampleData {
		noOfAllRequests = 750
		Eps = 1.2 //3.0
		MinSamples = 2.0
		X = generateSampleData(noOfAllRequests)
	}

	db := cluster.NewDBSCAN(&cluster.DBSCANConfig{Eps: Eps, MinSamples: MinSamples, Algorithm: ""})
	db.Fit(X, nil)
	coreSampleMask := make([]bool, len(db.Labels))
	for sample := range db.CoreSampleIndices {
		coreSampleMask[sample] = true
	}
	labels := db.Labels
	labelsmap := make(map[int]int)
	for _, l := range labels {
		labelsmap[l] = l
	}
	nclusters := len(labelsmap)
	if _, ok := labelsmap[-1]; ok {
		nclusters--
	}
	log.Printf("Estimated number of clusters: %d\n", nclusters)

	err := PlotResults(labelsmap, noOfAllRequests, labels, nclusters, X, appendName)
	if err != nil {
		log.Fatalf("\n%+v", err)

	}
	err = findClusterMembers(labels, X)
	if err != nil {
		log.Fatalf("\n%+v", err)

	}
}

func getX(noOfAllRequests int, lengthOfEachRequest int, allRequests []float64) *mat.Dense {
	return mat.NewDense(noOfAllRequests, lengthOfEachRequest, allRequests)
}

func generateSampleData(noOfAllRequests int) *mat.Dense {
	// Generate sample data
	centers := mat.NewDense(3, 2, []float64{1, 1, -1, -1, 1, -1})
	X, _ := datasets.MakeBlobs(&datasets.MakeBlobsConfig{NSamples: noOfAllRequests, Centers: centers, ClusterStd: 0.1}) //RandomState: rand.New(rand.NewSource(0)),
	X, _ = preprocessing.NewStandardScaler().FitTransform(X, nil)

	return X
}

func PlotResults(labelsmap map[int]int, noOfAllRequests int, labels []int, nclusters int, X *mat.Dense, appendName string) error {
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

	err = pt.Save(6*vg.Inch, 4*vg.Inch, pngfile);
	if err != nil {
		return errors.Wrap(err, "error saving png")
	}

	return nil
}
