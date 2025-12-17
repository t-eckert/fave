package internal

import (
	"time"
)

type Bookmark struct {
	Url         string   `json:"url"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	CreatedAt   int64    `json:"created_at"`
	UpdatedAt   int64    `json:"updated_at"`
}

func NewBookmark(url, name, description string, tags []string) Bookmark {
	return Bookmark{
		Url:         url,
		Name:        name,
		Description: description,
		Tags:        tags,
		CreatedAt:   nowUnix(),
		UpdatedAt:   nowUnix(),
	}
}

func nowUnix() int64 {
	return time.Now().Unix()
}
