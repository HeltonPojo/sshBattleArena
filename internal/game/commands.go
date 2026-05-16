package game

import "strings"

type InputType int

const (
	InputKeystroke InputType = iota // normal-mode key press
	InputCommand                    // full command string (after Enter in command mode)
)

type InputEvent struct {
	PlayerID string
	Type     InputType
	Key      rune   // the typed character (InputKeystroke)
	Command  string // raw command text without leading ":" (InputCommand)
}

func ParseCommand(raw string) (string, bool) {
	cmd := strings.TrimSpace(strings.ToLower(raw))
	switch cmd {
	case "up", "down", "left", "right", "bomb":
		return cmd, true
	default:
		return "", false
	}
}
