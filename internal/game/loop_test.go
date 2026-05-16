package game

import (
	"sync"
	"testing"
)

func newTestLoop() (*GameLoop, *sync.Map) {
	snaps := &sync.Map{}
	broadcast := func(playerID string, snap Snapshot) {
		snaps.Store(playerID, snap)
	}
	gl := NewGameLoop(broadcast)
	return gl, snaps
}

func TestMovement(t *testing.T) {
	gl, _ := newTestLoop()
	gl.AddPlayer("p1")
	p := gl.world.Players["p1"]
	startX, startY := p.X, p.Y

	// Send a keystroke — player faces right by default (slot 0).
	gl.handleInput(InputEvent{PlayerID: "p1", Type: InputKeystroke, Key: 'a'})

	if p.X != startX+1 || p.Y != startY {
		t.Errorf("expected move right to (%d,%d), got (%d,%d)", startX+1, startY, p.X, p.Y)
	}
	// Trail should be at old position.
	cell := gl.world.Grid[startY][startX]
	if cell.Type != CellTrail || cell.Rune != 'a' {
		t.Errorf("expected trail 'a' at (%d,%d), got type=%d rune=%c", startX, startY, cell.Type, cell.Rune)
	}
}

func TestCollisionBlocks(t *testing.T) {
	gl, _ := newTestLoop()
	gl.AddPlayer("p1")
	p := gl.world.Players["p1"]

	// Place trail ahead.
	gl.world.Grid[p.Y][p.X+1] = Cell{Type: CellTrail, Rune: 'x', OwnerID: "p1"}

	oldX, oldY := p.X, p.Y
	gl.handleInput(InputEvent{PlayerID: "p1", Type: InputKeystroke, Key: 'z'})

	if p.X != oldX || p.Y != oldY {
		t.Error("player should be blocked by trail")
	}
}

func TestCommandModeDirection(t *testing.T) {
	gl, _ := newTestLoop()
	gl.AddPlayer("p1")
	p := gl.world.Players["p1"]

	// Enter command mode.
	gl.handleInput(InputEvent{PlayerID: "p1", Type: InputKeystroke, Key: ':'})
	if !p.InCommandMode {
		t.Fatal("expected command mode")
	}

	// Type "up" + enter.
	for _, ch := range "up" {
		gl.handleInput(InputEvent{PlayerID: "p1", Type: InputKeystroke, Key: ch})
	}
	gl.handleInput(InputEvent{PlayerID: "p1", Type: InputKeystroke, Key: '\n'})

	if p.InCommandMode {
		t.Error("should have exited command mode")
	}
	if p.Dir != DirUp {
		t.Errorf("expected DirUp, got %d", p.Dir)
	}
	if p.LastCommandResult != CmdAccepted {
		t.Errorf("expected CmdAccepted, got %d", p.LastCommandResult)
	}
}

func TestBombDetonation(t *testing.T) {
	gl, snaps := newTestLoop()
	gl.AddPlayer("p1")
	gl.AddPlayer("p2")
	p1 := gl.world.Players["p1"]
	p2 := gl.world.Players["p2"]

	// Place p2 within blast radius of p1.
	p2.X, p2.Y = p1.X+1, p1.Y

	// Plant bomb via command.
	gl.executeCommand(p1, "bomb")

	if len(gl.world.Bombs) != 1 {
		t.Fatalf("expected 1 bomb, got %d", len(gl.world.Bombs))
	}

	// Move p1 out of blast radius.
	p1.X = p1.X + BombRadius + 2

	// Tick down the fuse.
	for i := 0; i < BombFuse; i++ {
		gl.tick()
	}

	// The broadcast should show p2 as the loser and p1 as the winner.
	snap, ok := snaps.Load("p2")
	if !ok {
		t.Fatal("expected snapshot for p2")
	}
	s := snap.(Snapshot)
	if s.Loser != "p2" {
		t.Errorf("expected loser=p2, got %s", s.Loser)
	}
	if s.Winner != "p1" {
		t.Errorf("expected winner=p1, got %s", s.Winner)
	}

	// After reset, p2 should be removed from the world.
	if _, exists := gl.world.Players["p2"]; exists {
		t.Error("p2 should be removed from world after reset")
	}
	// Winner should be respawned.
	if !gl.world.Players["p1"].Alive {
		t.Error("p1 should be alive after reset")
	}
}

func TestBombCooldown(t *testing.T) {
	gl, _ := newTestLoop()
	gl.AddPlayer("p1")
	p1 := gl.world.Players["p1"]

	gl.executeCommand(p1, "bomb")
	if p1.BombCooldownLeft != BombCooldown {
		t.Errorf("expected cooldown %d, got %d", BombCooldown, p1.BombCooldownLeft)
	}

	// Second bomb should be rejected.
	gl.executeCommand(p1, "bomb")
	if p1.LastCommandResult != CmdRejected {
		t.Error("expected bomb to be rejected during cooldown")
	}
	if len(gl.world.Bombs) != 1 {
		t.Errorf("expected 1 bomb, got %d", len(gl.world.Bombs))
	}
}
