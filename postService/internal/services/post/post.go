package post

import (
	"context"
	"github.com/yourusername/deepLink/post-service/internal/domain/models"
	"log/slog"
)

type Post struct {
	log            *slog.Logger
	pstCreator     PostCreator
	pstProvider    PostProvider
	pstModerator   PostModerator
	pstInteraction PostInteraction
	appProvider    AppProvider
}

type PostCreator interface {
	CreatePost(
		ctx context.Context,
		userID int64,
		content string,
		mediaURLs []string,
	) (postID int64, err error)
}

type PostProvider interface {
	GetPost(ctx context.Context, postID int64) (*models.Post, error)
	ListPosts(ctx context.Context, page, pageSize int32) ([]*models.Post, error)
}

type PostModerator interface {
	DeletePost(ctx context.Context, postID int64) (bool, error)
}

type PostInteraction interface {
	LikePost(ctx context.Context, userID, postID int64) (bool, error)
	AddComment(ctx context.Context, postID, userID int64, content string) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int32) (models.App, error)
}

// New returns a new instance of the Post service.
func New(
	log *slog.Logger,
	postCreator PostCreator,
	postProvider PostProvider,
	postModerator PostModerator,
	postInteraction PostInteraction,
	appProvider AppProvider,
) *Post {
	return &Post{
		log:            log,
		pstCreator:     postCreator,
		pstProvider:    postProvider,
		pstModerator:   postModerator,
		pstInteraction: postInteraction,
		appProvider:    appProvider,
	}
}
