package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/HeltoPojo/sshBattleArena/internal/game"
)

type quitMsg struct{}

type Model struct {
	PlayerID     string
	InputCh      chan<- game.InputEvent
	LastSnapshot game.Snapshot
	quitting     bool
}

func NewModel(playerID string, inputCh chan<- game.InputEvent) Model {
	return Model{
		PlayerID: playerID,
		InputCh:  inputCh,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case game.Snapshot:
		m.LastSnapshot = msg
		// Only the loser disconnects.
		if msg.Loser == m.PlayerID && !m.quitting {
			m.quitting = true
			return m, tea.Tick(2*time.Second, func(time.Time) tea.Msg {
				return quitMsg{}
			})
		}
		return m, nil

	case quitMsg:
		return m, tea.Quit

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.quitting {
			return m, nil
		}
		m.forwardKey(msg)
		return m, nil
	}
	return m, nil
}

func (m Model) View() string {
	snap := m.LastSnapshot
	if snap.Players == nil {
		return renderWaiting()
	}

	var s string
	s += renderGrid(snap, m.PlayerID)
	s += renderStatusBar(snap, m.PlayerID)

	if snap.Loser != "" {
		s += renderGameOver(snap, m.PlayerID)
	} else if snap.Waiting {
		s += "\n" + renderWaiting()
	}

	return s
}

func (m Model) forwardKey(msg tea.KeyMsg) {
	if m.InputCh == nil {
		return
	}

	ev := game.InputEvent{PlayerID: m.PlayerID}

	keyStr := msg.String()
	runes := []rune(keyStr)

	switch {
	case keyStr == "enter":
		ev.Type = game.InputKeystroke
		ev.Key = '\n'
	case keyStr == "backspace":
		ev.Type = game.InputKeystroke
		ev.Key = 127
	case len(runes) == 1:
		ev.Type = game.InputKeystroke
		ev.Key = runes[0]
	default:
		return
	}

	select {
	case m.InputCh <- ev:
	default:
	}
}
