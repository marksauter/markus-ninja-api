package resolver

type textMatchResolver struct {
}

func (r *textMatchResolver) Fragment() string {
	return ""
}

func (r *textMatchResolver) Highlights() []*textMatchHighlightResolver {
	var highlights []*textMatchHighlightResolver
	return highlights
}

func (r *textMatchResolver) Property() string {
	return ""
}
