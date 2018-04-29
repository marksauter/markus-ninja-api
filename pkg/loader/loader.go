package loader

import (
	"github.com/graph-gophers/dataloader"
)

func createLoader(batchFn dataloader.BatchFunc) *dataloader.Loader {
	return dataloader.NewBatchedLoader(batchFn)
}
