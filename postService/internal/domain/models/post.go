package models

import "time"

type Post struct {
	ID         int64
	UserID     int64
	Content    string
	MediaURLs  []string
	CreatedAt  time.Time
	UpdatedAt  *time.Time
	LikesCount int32
	Comments   []Comment
	IsDeleted  bool
	DeletedAt  *time.Time
}

type Comment struct {
	ID        int64
	UserID    int64
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Attachment struct {
	Type      string
	URL       string
	Thumbnail string
	Size      int64
	CreatedAt time.Time
}
