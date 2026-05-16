package game

import (
	"context"
	"sync"
	"time"
	"unicode"
)

type BroadcastFunc func(playerID string, snap Snapshot)

type GameLoop struct {
	mu        sync.Mutex
	world     *World
	inputCh   chan InputEvent
	broadcast BroadcastFunc
	nextSlot  int
}

func NewGameLoop(broadcast BroadcastFunc) *GameLoop {
	return &GameLoop{
		world:     NewWorld(),
		inputCh:   make(chan InputEvent, 256),
		broadcast: broadcast,
	}
}

func (gl *GameLoop) InputCh() chan<- InputEvent {
	return gl.inputCh
}

func (gl *GameLoop) AddPlayer(id string) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.world.SpawnPlayer(id, gl.nextSlot)
	gl.nextSlot++
}

func (gl *GameLoop) RemovePlayer(id string) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	if p, ok := gl.world.Players[id]; ok {
		p.Alive = false
	}
}

func (gl *GameLoop) ResetIfEmpty() {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	for _, p := range gl.world.Players {
		if p.Alive {
			return
		}
	}
	// All players dead or gone — reset.
	gl.world = NewWorld()
	gl.nextSlot = 0
}

func (gl *GameLoop) Run(ctx context.Context) {
	ticker := time.NewTicker(TickRate)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case ev := <-gl.inputCh:
			gl.mu.Lock()
			gl.handleInput(ev)
			gl.mu.Unlock()
		case <-ticker.C:
			gl.mu.Lock()
			gl.tick()
			gl.mu.Unlock()
		}
	}
}

func (gl *GameLoop) tick() {
	w := gl.world
	w.Tick++

	for _, p := range w.Players {
		if p.BombCooldownLeft > 0 {
			p.BombCooldownLeft--
		}
		// Clear stale command results.
		if p.LastCommandResult != CmdNone && w.Tick-p.LastCommandTick >= CommandResultTTL {
			p.LastCommandResult = CmdNone
		}
	}

	for y := range GridHeight {
		for x := range GridWidth {
			cell := &w.Grid[y][x]
			if cell.Type == CellExplosion {
				cell.TTL--
				if cell.TTL <= 0 {
					*cell = Cell{}
				}
			}
		}
	}

	remaining := w.Bombs[:0]
	for i := range w.Bombs {
		b := &w.Bombs[i]
		b.FuseLeft--
		if b.FuseLeft <= 0 {
			gl.detonate(b)
		} else {
			remaining = append(remaining, *b)
		}
	}
	w.Bombs = remaining

	var winnerID, loserID string
	alive := 0
	for _, p := range w.Players {
		if p.Alive {
			alive++
			winnerID = p.ID
		} else {
			loserID = p.ID
		}
	}
	roundOver := len(w.Players) == 2 && alive == 1 && loserID != ""

	if roundOver {
		for id := range w.Players {
			snap := w.TakeSnapshot(id)
			snap.Loser = loserID
			snap.Winner = winnerID
			gl.broadcast(id, snap)
		}
		w.ResetForWinner(winnerID)
		gl.nextSlot = 1 // winner is at slot 0, next joiner gets slot 1
	} else {
		for id := range w.Players {
			snap := w.TakeSnapshot(id)
			gl.broadcast(id, snap)
		}
	}
}

func (gl *GameLoop) detonate(b *Bomb) {
	w := gl.world
	for dy := -b.Radius; dy <= b.Radius; dy++ {
		for dx := -b.Radius; dx <= b.Radius; dx++ {
			ny, nx := b.Y+dy, b.X+dx
			if ny < 0 || ny >= GridHeight || nx < 0 || nx >= GridWidth {
				continue
			}
			cell := &w.Grid[ny][nx]
			switch cell.Type {
			case CellTrail, CellBomb:
				*cell = Cell{Type: CellExplosion, TTL: ExplosionTTL}
			case CellEmpty:
				*cell = Cell{Type: CellExplosion, TTL: ExplosionTTL}
			}
		}
	}
	for _, p := range w.Players {
		if !p.Alive {
			continue
		}
		dx := p.X - b.X
		dy := p.Y - b.Y
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}
		if dx <= b.Radius && dy <= b.Radius {
			p.Alive = false
		}
	}
}

func (gl *GameLoop) handleInput(ev InputEvent) {
	w := gl.world
	p, ok := w.Players[ev.PlayerID]
	if !ok || !p.Alive {
		return
	}

	// Block all input while waiting for a second player.
	alive := 0
	for _, pl := range w.Players {
		if pl.Alive {
			alive++
		}
	}
	if alive < 2 {
		return
	}

	switch ev.Type {
	case InputKeystroke:
		gl.handleKeystroke(p, ev.Key)
	case InputCommand:
		gl.handleCommand(p, ev.Command)
	}
}

func (gl *GameLoop) handleKeystroke(p *Player, key rune) {
	w := gl.world

	if key == ':' && !p.InCommandMode {
		p.InCommandMode = true
		p.CommandBuf = ""
		return
	}

	if p.InCommandMode {
		switch key {
		case '\n', '\r': // Enter — execute command
			cmd, valid := ParseCommand(p.CommandBuf)
			if valid {
				gl.executeCommand(p, cmd)
				p.LastCommandResult = CmdAccepted
			} else {
				p.LastCommandResult = CmdRejected
			}
			p.LastCommandTick = w.Tick
			p.InCommandMode = false
			p.CommandBuf = ""
		case 127, '\b': // Backspace
			if len(p.CommandBuf) > 0 {
				p.CommandBuf = p.CommandBuf[:len(p.CommandBuf)-1]
			}
		default:
			if unicode.IsPrint(key) {
				p.CommandBuf += string(key)
			}
		}
		return
	}

	if !unicode.IsPrint(key) || key == ' ' {
		return
	}

	dx, dy := dirDelta(p.Dir)
	nx, ny := p.X+dx, p.Y+dy
	if nx < 0 || nx >= GridWidth || ny < 0 || ny >= GridHeight {
		return
	}
	target := &w.Grid[ny][nx]
	if target.Type != CellEmpty && target.Type != CellExplosion {
		return
	}

	w.Grid[p.Y][p.X] = Cell{Type: CellTrail, Rune: key, OwnerID: p.ID}
	p.X, p.Y = nx, ny
}

func (gl *GameLoop) handleCommand(p *Player, command string) {
	cmd, valid := ParseCommand(command)
	if valid {
		gl.executeCommand(p, cmd)
		p.LastCommandResult = CmdAccepted
	} else {
		p.LastCommandResult = CmdRejected
	}
	p.LastCommandTick = gl.world.Tick
}

func (gl *GameLoop) executeCommand(p *Player, cmd string) {
	w := gl.world
	switch cmd {
	case "up":
		p.Dir = DirUp
	case "down":
		p.Dir = DirDown
	case "left":
		p.Dir = DirLeft
	case "right":
		p.Dir = DirRight
	case "bomb":
		if p.BombCooldownLeft > 0 {
			p.LastCommandResult = CmdRejected
			p.LastCommandTick = w.Tick
			return
		}
		bomb := Bomb{
			X: p.X, Y: p.Y,
			FuseLeft: BombFuse,
			OwnerID:  p.ID,
			Radius:   BombRadius,
		}
		w.Bombs = append(w.Bombs, bomb)
		w.Grid[p.Y][p.X] = Cell{Type: CellBomb, OwnerID: p.ID}
		p.BombCooldownLeft = BombCooldown
	}
}

func dirDelta(d Direction) (dx, dy int) {
	switch d {
	case DirUp:
		return 0, -1
	case DirDown:
		return 0, 1
	case DirLeft:
		return -1, 0
	case DirRight:
		return 1, 0
	default:
		return 0, 0
	}
}
