package models

type PinnedChat struct {
	ChatId string `json:"chat_id"`
	Pinned bool   `json:"pinned"`
}
