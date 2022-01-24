package store

type Feed struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type FeedStore interface {
	List() ([]Feed, error)
	Insert(Feed) (uint, error)
}
