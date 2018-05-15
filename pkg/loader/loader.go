package loader

import (
	"github.com/graph-gophers/dataloader"
)

type Loader interface {
	Clear(string)
	ClearAll()
}

func createLoader(batchFn dataloader.BatchFunc) *dataloader.Loader {
	return dataloader.NewBatchedLoader(batchFn)
}
