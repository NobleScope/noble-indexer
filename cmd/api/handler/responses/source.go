package responses

import (
	"github.com/baking-bad/noble-indexer/internal/storage"
)

// Source model info
//
//	@Description	Contract source information
type Source struct {
	Id      uint64 `example:"321"                  json:"id"      swaggertype:"integer"`
	Name    string `example:"Contract source name" json:"name"    swaggertype:"string"`
	License string `example:"License"              json:"license" swaggertype:"string"`
	Content string `example:"Source content"       json:"content" swaggertype:"string"`

	Urls []string `json:"urls,omitempty"`
}

func NewSource(source storage.Source) Source {
	s := Source{
		Id:      source.Id,
		Name:    source.Name,
		License: source.License,
		Content: source.Content,
		Urls:    source.Urls,
	}

	return s
}
