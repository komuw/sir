package sir

import (
	"log"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat/mds"
)

func FindMDS() {
	/*
		Multidimensional scaling (MDS) seeks a low-dimensional representation of the data in which the distances respect well the distances in the original high-dimensional space.
	*/

	// The dis matrix must be square or TorgersonScaling will panic.
	dis := mat.NewSymDense(8, []float64{
		0, 1328, 1600, 2616, 1161, 653, 2130, 1161,
		1328, 0, 1962, 1289, 2463, 1889, 1991, 2026,
		1600, 1962, 0, 2846, 1788, 1374, 3604, 732,
		2616, 1289, 2846, 0, 3734, 3146, 2652, 3146,
		1161, 2463, 1788, 3734, 0, 598, 3008, 1057,
		653, 1889, 1374, 3146, 598, 0, 2720, 713,
		2130, 1991, 3604, 2652, 3008, 2720, 0, 3288,
		1161, 2026, 732, 3146, 1057, 713, 3288, 0,
	})
	_, c := dis.Dims()
	k, mad, eig := mds.TorgersonScaling(nil, make([]float64, c), dis)

	log.Println("k", k)
	log.Println("eig", eig)
	log.Println("mad", mad)

}
