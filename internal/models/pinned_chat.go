package models

type PinnedChat struct {
	ChatId string `json:"chat_id"`
	Pinned bool   `json:"pinned"`
}

type PinChatRequest struct {
	ChatID string `json:"chat_id"`
	Pin    bool   `json:"pin"`
}
