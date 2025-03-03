package models

type (
	Background struct {
		ID       string `bson:"_id,omitempty"`
		Filename string `bson:"filename"`
	}
)
