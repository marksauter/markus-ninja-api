package repo

import "github.com/marksauter/markus-ninja-api/pkg/model"

type Repo interface {
	Get(id string) model.Node
}
