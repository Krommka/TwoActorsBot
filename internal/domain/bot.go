package domain

import ()

type BotHandler interface {
	HandleCommand(chatID int64, command string, query string) (string, error)
	SendMessage(chatID int64, text string) error
	SendPhoto(chatID int64, photoURL string, caption string) error
}
