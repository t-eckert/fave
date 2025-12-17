package internal

type Bookmark struct {
	Url         string   `json:"url"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}
