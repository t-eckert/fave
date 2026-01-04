package internal

import (
	"time"
)

// Bookmark represents a single bookmark with metadata.
type Bookmark struct {
	Url         string   `json:"url" yaml:"url"`
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description" yaml:"description"`
	Tags        []string `json:"tags" yaml:"tags"`
	CreatedAt   int64    `json:"created_at" yaml:"created_at"`
	UpdatedAt   int64    `json:"updated_at" yaml:"updated_at"`
}

// NewBookmark creates a new Bookmark instance with the provided details.
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
