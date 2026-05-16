package server

import "github.com/HeltoPojo/sshBattleArena/internal/game"

func NewBroadcaster(reg *Registry) game.BroadcastFunc {
	return func(playerID string, snap game.Snapshot) {
		p := reg.GetProgram(playerID)
		if p == nil {
			return
		}
		p.Send(snap)
	}
}
