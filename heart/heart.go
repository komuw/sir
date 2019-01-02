package heart

import (
	"github.com/pa-m/sklearn/cluster"
	"github.com/pa-m/sklearn/datasets"
	"github.com/pa-m/sklearn/preprocessing"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/mat"
)

/*
usage:
  go run -race go_dbscan/go_dbscan.go
*/

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
	_, _ = rows, columns
	for k := range labels {
		for i := 0; i < columns; i++ {
			zAt := X.At(k, i)
			_ = zAt
			// log.Println("zAt", zAt)
		}

	}

	// TODO: after establishing the members of the clusters, that's where we ought to call
	// PlotHeatMap
	return nil

}

func GetClusters(noOfAllRequests int, lengthOfLargestRequest int, allRequests []float64, Eps float64, MinSamples float64, autoGenerateSampleData bool, appendName string) (int, *mat.Dense, error) {
	// adapted from http://scikit-learn.org/stable/_downloads/plot_dbscan.ipynb
	if lengthOfLargestRequest <= 1 {
		err := errors.New("we cant create a matrix with no dimensions, ie X.At(x, y) will fail")
		return 0, nil, err
	}

	X := getX(noOfAllRequests, lengthOfLargestRequest, allRequests)
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
	return nclusters, X, nil
}

func getX(noOfAllRequests int, lengthOfLargestRequest int, allRequests []float64) *mat.Dense {
	return mat.NewDense(noOfAllRequests, lengthOfLargestRequest, allRequests)
}

func generateSampleData(noOfAllRequests int) *mat.Dense {
	// Generate sample data
	centers := mat.NewDense(3, 2, []float64{1, 1, -1, -1, 1, -1})
	X, _ := datasets.MakeBlobs(&datasets.MakeBlobsConfig{NSamples: noOfAllRequests, Centers: centers, ClusterStd: 0.1}) //RandomState: rand.New(rand.NewSource(0)),
	X, _ = preprocessing.NewStandardScaler().FitTransform(X, nil)

	return X
}
