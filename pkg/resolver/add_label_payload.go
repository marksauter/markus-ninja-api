package resolver

import (
	"fmt"

	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type AddLabelPayload = addLabelPayloadResolver

type addLabelPayloadResolver struct {
	Label   *repo.LabelPermit
	Labeled *repo.LabeledPermit
	Repos   *repo.Repos
}

func (r *addLabelPayloadResolver) LabelEdge() (*labelEdgeResolver, error) {
	return NewLabelEdgeResolver(r.Label, r.Repos)
}

func (r *addLabelPayloadResolver) Labelable() (*labelableResolver, error) {
	labelableId, err := r.Labeled.LabelableId()
	if err != nil {
		return nil, err
	}
	switch labelableId.Type {
	case "Lesson":
		lesson, err := r.Repos.Lesson().Get(labelableId.String)
		if err != nil {
			return nil, err
		}
		return &labelableResolver{&lessonResolver{Lesson: lesson, Repos: r.Repos}}, nil
	default:
		return nil, fmt.Errorf("invalid type '%s' for labeled labelable id", labelableId.String)
	}

}
