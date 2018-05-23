package loader

import (
	"strings"

	"github.com/graph-gophers/dataloader"
)

type Loader interface {
	Clear(string)
	ClearAll()
}

func createLoader(batchFn dataloader.BatchFunc) *dataloader.Loader {
	return dataloader.NewBatchedLoader(batchFn)
}

func newCompositeKey(strs ...string) dataloader.Key {
	return dataloader.StringKey(strings.Join(strs, ":"))
}

func splitCompositeKey(k dataloader.Key) []string {
	return strings.Split(k.String(), ":")
}
