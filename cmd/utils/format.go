package utils

import (
	"fmt"
	"time"

	"github.com/t-eckert/fave/internal"
)

func FormatDate(unixTime int64) string {
	t := time.Unix(unixTime, 0)
	return t.Format("2006-01-02 15:04:05")
}

func FormatBookmark(id int, bookmark *internal.Bookmark) string {
	return fmt.Sprintf("ID: %d\nName: %s\nURL: %s\nDescription: %s\nTags: %v\nCreated At: %s\nUpdated At: %s",
		id,
		bookmark.Name,
		bookmark.Url,
		bookmark.Description,
		bookmark.Tags,
		FormatDate(bookmark.CreatedAt),
		FormatDate(bookmark.UpdatedAt))
}
