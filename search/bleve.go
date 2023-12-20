package search

import (
	"github.com/blevesearch/bleve/v2"
)

func CreateIndex() (bleve.Index, error) {
	open, err := bleve.Open("caddy.bleve")
	if err != nil {
		if err != bleve.ErrorIndexPathDoesNotExist {
			return nil, err
		}
		open, err = bleve.New("caddy.bleve", bleve.NewIndexMapping())
		if err != nil {
			return nil, err
		}
	}
	return open, nil
}
