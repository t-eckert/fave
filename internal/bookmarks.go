package internal

import (
	"sync"
)

type Bookmark struct {
	Url         string   `json:"url"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

var Bookmarks = map[int]Bookmark{}

var BookmarksMutex = new(sync.RWMutex)
