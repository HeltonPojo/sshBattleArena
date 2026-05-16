package tui

import (
	"fmt"
	"strings"

	"github.com/HeltoPojo/sshBattleArena/internal/game"
	"github.com/charmbracelet/lipgloss"
)

var (
	borderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	p1Style     = lipgloss.NewStyle().Foreground(lipgloss.Color("6")) // cyan
	p2Style     = lipgloss.NewStyle().Foreground(lipgloss.Color("5")) // magenta
	bombStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // red
	bombWarn    = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	cmdStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // yellow
	acceptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // green
	rejectStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // red
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func renderGrid(snap game.Snapshot, myID string) string {
	playerPos := make(map[[2]int]*game.Player)
	playerOrder := buildPlayerOrder(snap, myID)
	for i := range playerOrder {
		p := &playerOrder[i]
		if p.Alive {
			playerPos[[2]int{p.X, p.Y}] = p
		}
	}

	bombPos := make(map[[2]int]*game.Bomb)
	for i := range snap.Bombs {
		b := &snap.Bombs[i]
		bombPos[[2]int{b.X, b.Y}] = b
	}

	var sb strings.Builder
	for y := range game.GridHeight {
		for x := range game.GridWidth {
			pos := [2]int{x, y}

			if p, ok := playerPos[pos]; ok {
				glyph := string(game.DirGlyph(p.Dir))
				style := playerStyle(p.ID, playerOrder, myID)
				sb.WriteString(style.Render(glyph))
				continue
			}

			if b, ok := bombPos[pos]; ok {
				digit := fmt.Sprintf("%d", b.FuseLeft)
				if b.FuseLeft <= 5 {
					sb.WriteString(bombWarn.Render(digit))
				} else {
					sb.WriteString(bombStyle.Render(digit))
				}
				continue
			}

			cell := snap.Grid[y][x]
			switch cell.Type {
			case game.CellBorder:
				sb.WriteString(borderStyle.Render("#"))
			case game.CellTrail:
				style := trailStyle(cell.OwnerID, playerOrder, myID)
				sb.WriteString(style.Render(string(cell.Rune)))
			case game.CellBomb:
				sb.WriteString(bombStyle.Render("*"))
			default:
				sb.WriteString(" ")
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func renderStatusBar(snap game.Snapshot, myID string) string {
	me, ok := snap.Players[myID]
	if !ok {
		return ""
	}

	var parts []string

	parts = append(parts, fmt.Sprintf("Dir: %s", dirName(me.Dir)))

	// Bomb cooldown.
	if me.BombCooldownLeft > 0 {
		parts = append(parts, fmt.Sprintf("Bomb: %dt", me.BombCooldownLeft))
	} else {
		parts = append(parts, acceptStyle.Render("Bomb: ready"))
	}

	// Command mode indicator.
	if me.InCommandMode {
		parts = append(parts, cmdStyle.Render(fmt.Sprintf("CMD> :%s", me.CommandBuf)))
	}

	// Command result flash.
	if me.LastCommandResult == game.CmdAccepted {
		parts = append(parts, acceptStyle.Render("[OK]"))
	} else if me.LastCommandResult == game.CmdRejected {
		parts = append(parts, rejectStyle.Render("[INVALID]"))
	}

	return strings.Join(parts, "  ")
}

func renderGameOver(snap game.Snapshot, myID string) string {
	if snap.Winner == myID {
		return acceptStyle.Bold(true).Render("\n  YOU WIN!  \n")
	}
	return rejectStyle.Bold(true).Render("\n  YOU LOST!  \n")
}

func renderWaiting() string {
	return dimStyle.Render("Waiting for opponent...\n")
}

func buildPlayerOrder(snap game.Snapshot, myID string) []game.Player {
	players := make([]game.Player, 0, len(snap.Players))
	if me, ok := snap.Players[myID]; ok {
		players = append(players, me)
	}
	for id, p := range snap.Players {
		if id != myID {
			players = append(players, p)
		}
	}
	return players
}

func playerStyle(id string, _ []game.Player, myID string) lipgloss.Style {
	if id == myID {
		return p1Style
	}
	return p2Style
}

func trailStyle(ownerID string, _ []game.Player, myID string) lipgloss.Style {
	if ownerID == myID {
		return p1Style
	}
	return p2Style
}

func dirName(d game.Direction) string {
	switch d {
	case game.DirUp:
		return "UP"
	case game.DirDown:
		return "DOWN"
	case game.DirLeft:
		return "LEFT"
	case game.DirRight:
		return "RIGHT"
	default:
		return "?"
	}
}
