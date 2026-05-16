package game

import "time"

// Game constants — tweak these to retune gameplay.
const (
	GridWidth        = 100
	GridHeight       = 50
	TickRate         = 100 * time.Millisecond // ~10 ticks/sec
	BombFuse         = 60                     // ticks
	BombCooldown     = 100                    // ticks
	BombRadius       = 5
	CommandResultTTL = 5 // ticks before result clears
)

type World struct {
	Grid    [GridHeight][GridWidth]Cell
	Players map[string]*Player
	Bombs   []Bomb
	Tick    int
}

func NewWorld() *World {
	w := &World{
		Players: make(map[string]*Player),
	}
	for y := 0; y < GridHeight; y++ {
		for x := 0; x < GridWidth; x++ {
			if y == 0 || y == GridHeight-1 || x == 0 || x == GridWidth-1 {
				w.Grid[y][x] = Cell{Type: CellBorder}
			}
		}
	}
	return w
}

// SpawnPlayer places a player in one of the two starting slots.
func (w *World) SpawnPlayer(id string, slot int) {
	p := &Player{ID: id, Alive: true}
	switch slot {
	case 0: // top-left, facing right
		p.X, p.Y, p.Dir = 2, 2, DirRight
	default: // bottom-right, facing left
		p.X, p.Y, p.Dir = GridWidth-3, GridHeight-3, DirLeft
	}
	w.Players[id] = p
}

// Snapshot is a deep-copy, read-only view of the world sent to renderers.
type Snapshot struct {
	Grid    [GridHeight][GridWidth]Cell
	Players map[string]Player // value copies
	Bombs   []Bomb
	Tick    int
	Loser   string // set to the dead player's ID on kill; winner stays connected
	Winner  string // set to the surviving player's ID on kill
}

func (w *World) ResetForWinner(winnerID string) {
	// Clear grid to borders + empty.
	for y := range GridHeight {
		for x := range GridWidth {
			if y == 0 || y == GridHeight-1 || x == 0 || x == GridWidth-1 {
				w.Grid[y][x] = Cell{Type: CellBorder}
			} else {
				w.Grid[y][x] = Cell{}
			}
		}
	}
	w.Bombs = nil

	// Remove dead players, respawn winner.
	for id := range w.Players {
		if id != winnerID {
			delete(w.Players, id)
		}
	}
	if p, ok := w.Players[winnerID]; ok {
		p.X, p.Y, p.Dir = 2, 2, DirRight
		p.Alive = true
		p.BombCooldownLeft = 0
		p.InCommandMode = false
		p.CommandBuf = ""
		p.LastCommandResult = CmdNone
	}
}

// TakeSnapshot produces a snapshot for a specific player's perspective.
func (w *World) TakeSnapshot(forPlayerID string) Snapshot {
	s := Snapshot{
		Grid:    w.Grid,
		Players: make(map[string]Player, len(w.Players)),
		Tick:    w.Tick,
	}
	for id, p := range w.Players {
		s.Players[id] = *p
	}
	s.Bombs = make([]Bomb, len(w.Bombs))
	copy(s.Bombs, w.Bombs)
	return s
}
