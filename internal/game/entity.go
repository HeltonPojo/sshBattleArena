package game

type Direction int

const (
	DirUp Direction = iota
	DirDown
	DirLeft
	DirRight
)

func DirGlyph(d Direction) rune {
	switch d {
	case DirUp:
		return '▲'
	case DirDown:
		return '▼'
	case DirLeft:
		return '◀'
	case DirRight:
		return '▶'
	default:
		return '?'
	}
}

func DirGlyphASCII(d Direction) rune {
	switch d {
	case DirUp:
		return '^'
	case DirDown:
		return 'v'
	case DirLeft:
		return '<'
	case DirRight:
		return '>'
	default:
		return '?'
	}
}

type CellType int

const (
	CellEmpty  CellType = iota
	CellBorder          // indestructible wall
	CellTrail           // letter trail (destructible)
	CellBomb            // active bomb
	CellExplosion       // explosion animation (short-lived)
)

type Cell struct {
	Type    CellType
	Rune    rune   // the letter for CellTrail, 0 otherwise
	OwnerID string // player who placed it (trail or bomb)
	TTL     int    // ticks remaining for CellExplosion
}

type CommandResult int

const (
	CmdNone     CommandResult = iota
	CmdAccepted               // green flash
	CmdRejected               // red flash
)

type Player struct {
	ID                string
	X, Y              int
	Dir               Direction
	Alive             bool
	BombCooldownLeft  int
	InCommandMode     bool
	CommandBuf        string
	LastCommandResult CommandResult
	LastCommandTick   int
}

type Bomb struct {
	X, Y     int
	FuseLeft int
	OwnerID  string
	Radius   int
}
