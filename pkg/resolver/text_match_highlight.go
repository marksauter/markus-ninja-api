package resolver

type textMatchHighlightResolver struct {
	Begin int32
	End   int32
}

func (r *textMatchHighlightResolver) BeginIndice() int32 {
	return r.Begin
}

func (r *textMatchHighlightResolver) EndIndice() int32 {
	return r.End
}

func (r *textMatchHighlightResolver) Text() string {
	return ""
}
