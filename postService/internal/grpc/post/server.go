package post

import (
	"context"
	postv1 "github.com/NorthDice/DeepLink/protos/gen/go/post"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type postService interface {
	CreatePost(
		ctx context.Context,
		userId int64,
		content string,
		mediaURLs []string,
	) (int64, error)
	GetPost(
		ctx context.Context,
		postId int64,
	) (*Post, error)
	ListPost(
		ctx context.Context,
		page int32,
		pageSize int32,
	) ([]*Post, error)
	DeletePost(
		ctx context.Context,
		postId int64,
	) (bool, error)
	LikePost(
		ctx context.Context,
		postID,
		userID int64,
	) (bool, error)
	AddComment(
		ctx context.Context,
		postID,
		userID int64,
		content string,
	) (bool, error)
}

type serverAPI struct {
	postv1.UnimplementedPostServer
	service postService
}

const (
	emptyValue  = 0
	emptyString = ""
)

func Register(gRPC *grpc.Server) {
	postv1.RegisterPostServer(gRPC, &serverAPI{})
}

func (s *serverAPI) CreatePost(
	ctx context.Context,
	req *postv1.CreatePostRequest,
) (*postv1.CreatePostResponse, error) {

	if err := validateCreatePost(req); err != nil {
		return nil, err
	}

	postID, err := s.service.CreatePost(ctx, req.GetUserId(), req.GetContent(), req.GetMediaUrl())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create post")
	}

	return &postv1.CreatePostResponse{PostId: postID}, nil
}

func (s *serverAPI) GetPost(
	ctx context.Context,
	req *postv1.GetPostRequest,
) (*postv1.GetPostResponse, error) {

	if err := validateGetPost(req); err != nil {
		return nil, err
	}

	post, err := s.service.GetPost(ctx, req.GetPostId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "post not found")
	}

	return convertPostToProto(post), nil
}

func (s *serverAPI) ListPosts(
	ctx context.Context,
	req *postv1.ListPostsRequest,
) (*postv1.ListPostsResponse, error) {

	if err := validateListPosts(req); err != nil {
		return nil, err
	}

	posts, err := s.service.ListPost(ctx, req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list posts")
	}

	return &postv1.ListPostsResponse{
		Posts: convertPostsToProto(posts),
	}, nil
}

func (s *serverAPI) DeletePost(
	ctx context.Context,
	req *postv1.DeletePostRequest,
) (*postv1.DeletePostResponse, error) {

	if err := validateDeletePost(req); err != nil {
		return nil, err
	}

	success, err := s.service.DeletePost(ctx, req.GetPostId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete post")
	}

	return &postv1.DeletePostResponse{Success: success}, nil

}

func (s *serverAPI) LikePost(
	ctx context.Context,
	req *postv1.LikePostRequest,
) (*postv1.LikePostResponse, error) {

	if err := validateLikePost(req); err != nil {
		return nil, err
	}

	success, err := s.service.LikePost(ctx, req.GetUserId(), req.GetPostId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to like post")
	}

	return &postv1.LikePostResponse{Success: success}, nil
}

func (s *serverAPI) AddComment(
	ctx context.Context,
	req *postv1.AddCommentRequest,
) (*postv1.AddCommentResponse, error) {

	if err := validateAddComment(req); err != nil {
		return nil, err
	}

	success, err := s.service.AddComment(ctx, req.GetPostId(), req.GetUserId(), req.GetContent())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to add comment")
	}

	return &postv1.AddCommentResponse{Success: success}, nil

}

type Post struct {
	ID         int64
	UserID     int64
	Content    string
	MediaURLs  []string
	CreatedAt  time.Time
	LikesCount int32
	Comments   []Comment
}

type Comment struct {
	UserID    int64
	Content   string
	CreatedAt time.Time
}

func validateCreatePost(req *postv1.CreatePostRequest) error {
	if req.GetUserId() == emptyValue {
		return status.Error(codes.InvalidArgument, "user id is required")
	}
	if req.GetContent() == emptyString {
		return status.Error(codes.InvalidArgument, "content is required")
	}
	return nil
}

func validateGetPost(req *postv1.GetPostRequest) error {
	if req.GetPostId() == emptyValue {
		return status.Error(codes.InvalidArgument, "postId is required")
	}
	return nil
}

func validateListPosts(req *postv1.ListPostsRequest) error {
	if req.GetPage() < 1 {
		return status.Error(codes.InvalidArgument, "page is required")
	}
	if req.GetPageSize() < 1 {
		return status.Error(codes.InvalidArgument, "pageSize is required")
	}

	return nil
}

func validateDeletePost(req *postv1.DeletePostRequest) error {
	if req.GetPostId() == emptyValue {
		return status.Error(codes.InvalidArgument, "postId is required")
	}

	return nil
}

func validateLikePost(req *postv1.LikePostRequest) error {

	if req.GetPostId() == emptyValue {
		return status.Error(codes.InvalidArgument, "postId is required")
	}
	if req.GetUserId() == emptyValue {
		return status.Error(codes.InvalidArgument, "userId is required")
	}

	return nil
}

func validateAddComment(req *postv1.AddCommentRequest) error {

	if req.GetPostId() == emptyValue {
		return status.Error(codes.InvalidArgument, "postId is required")
	}
	if req.GetUserId() == emptyValue {
		return status.Error(codes.InvalidArgument, "userId is required")
	}
	if req.GetContent() == emptyString {
		return status.Error(codes.InvalidArgument, "content is required")
	}

	return nil
}

func convertPostToProto(post *Post) *postv1.GetPostResponse {
	return &postv1.GetPostResponse{
		PostId:     post.ID,
		UserId:     post.UserID,
		Content:    post.Content,
		MediaUrl:   post.MediaURLs,
		CreatedAt:  timestamppb.New(post.CreatedAt),
		LikesCount: post.LikesCount,
		Comments:   convertCommentsToProto(post.Comments),
	}
}

func convertPostsToProto(posts []*Post) []*postv1.GetPostResponse {
	var result []*postv1.GetPostResponse
	for _, post := range posts {
		result = append(result, convertPostToProto(post))
	}
	return result
}

func convertCommentsToProto(comments []Comment) []*postv1.Comment {
	var result []*postv1.Comment
	for _, c := range comments {
		result = append(result, &postv1.Comment{
			UserId:    c.UserID,
			Content:   c.Content,
			CreatedAt: timestamppb.New(c.CreatedAt),
		})
	}
	return result
}
