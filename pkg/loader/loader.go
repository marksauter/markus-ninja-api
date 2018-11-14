package loader

import (
	"errors"
	"strings"

	"github.com/graph-gophers/dataloader"
)

var ErrWrongType = errors.New("wrong type")

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
