package resolver

import "github.com/marksauter/markus-ninja-api/pkg/repo"

type searchResultItemResolver struct {
	Item  repo.Permit
	Repos *repo.Repos
}

func (r *searchResultItemResolver) ToLesson() (*lessonResolver, bool) {
	lesson, ok := r.Item.(*repo.LessonPermit)
	return &lessonResolver{Lesson: lesson, Repos: r.Repos}, ok
}

func (r *searchResultItemResolver) ToStudy() (*studyResolver, bool) {
	study, ok := r.Item.(*repo.StudyPermit)
	return &studyResolver{Study: study, Repos: r.Repos}, ok
}

func (r *searchResultItemResolver) ToUser() (*userResolver, bool) {
	user, ok := r.Item.(*repo.UserPermit)
	return &userResolver{User: user, Repos: r.Repos}, ok
}

func (r *searchResultItemResolver) ToUserAsset() (*userAssetResolver, bool) {
	userAsset, ok := r.Item.(*repo.UserAssetPermit)
	return &userAssetResolver{UserAsset: userAsset, Repos: r.Repos}, ok
}
