package resolver

type textMatchHighlightResolver struct{}

func (r *textMatchHighlightResolver) BeginIndice() int32 {
	return 0
}

func (r *textMatchHighlightResolver) EndIndice() int32 {
	return 0
}

func (r *textMatchHighlightResolver) Text() string {
	return ""
}
