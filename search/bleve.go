package search

import (
	"fmt"
	"github.com/blevesearch/bleve"
)

func sd() {
	message := struct {
		Id   string
		From string
		Body string
	}{
		Id:   "example",
		From: "marty.schoch@gmail.com",
		Body: "bleve indexing is easy",
	}

	mapping := bleve.NewIndexMapping()
	index, err := bleve.New("example.bleve", mapping)
	if err != nil {
		index, err = bleve.Open("example.bleve")
	}
	index.Index(message.Id, message)

	//index, _ := bleve.Open("example.bleve")
	query := bleve.NewQueryStringQuery("bleve")
	searchRequest := bleve.NewSearchRequest(query)
	searchResult, _ := index.Search(searchRequest)
	fmt.Println(searchResult)
}
