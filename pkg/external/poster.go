package external

import "github.com/lamplitlabs/ferret/pkg/config"

type Post struct {
	Title       string
	Link        string
	Description string
	HashTags    string
}
type Poster interface {
	Post(configData config.Config, post Post) error
}
