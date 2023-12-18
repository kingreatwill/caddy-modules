package search

type Search struct {
	Root     string `json:"root,omitempty"`
	Endpoint string `json:"endpoint,omitempty"` // default: /search
	Regexp   string `json:"regexp,omitempty"`
}
