package resolver

import (
	"fmt"
	"strings"
)

type CommentableType int

const (
	CommentableTypeLessonComment CommentableType = iota
	CommentableTypeUserAssetComment
)

func ParseCommentableType(s string) (CommentableType, error) {
	switch strings.ToUpper(s) {
	case "LESSON_COMMENT":
		return CommentableTypeLessonComment, nil
	case "USER_ASSET_COMMENT":
		return CommentableTypeUserAssetComment, nil
	default:
		var f CommentableType
		return f, fmt.Errorf("invalid CommentableType: %q", s)
	}
}

func (f CommentableType) String() string {
	switch f {
	case CommentableTypeLessonComment:
		return "lesson_comment"
	case CommentableTypeUserAssetComment:
		return "user_asset_comment"
	default:
		return "unknown"
	}
}
