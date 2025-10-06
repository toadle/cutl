package internal

import tea "github.com/charmbracelet/bubbletea"

var (
	MessageRelay *relay
)

type relay struct {
	SendMsg func(msg tea.Msg)
}

func InitMessageRelay(sendMsg func(msg tea.Msg)) {
	MessageRelay = &relay{
		SendMsg: sendMsg,
	}
}
