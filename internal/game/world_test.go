package game

import "testing"

func TestNewWorldBorders(t *testing.T) {
	w := NewWorld()

	// Top and bottom rows should be border.
	for x := range GridWidth {
		if w.Grid[0][x].Type != CellBorder {
			t.Errorf("expected border at (0,%d), got %d", x, w.Grid[0][x].Type)
		}
		if w.Grid[GridHeight-1][x].Type != CellBorder {
			t.Errorf("expected border at (%d,%d), got %d", GridHeight-1, x, w.Grid[GridHeight-1][x].Type)
		}
	}

	// Left and right columns should be border.
	for y := range GridHeight {
		if w.Grid[y][0].Type != CellBorder {
			t.Errorf("expected border at (%d,0), got %d", y, w.Grid[y][0].Type)
		}
		if w.Grid[y][GridWidth-1].Type != CellBorder {
			t.Errorf("expected border at (%d,%d), got %d", y, GridWidth-1, w.Grid[y][GridWidth-1].Type)
		}
	}

	// Interior should be empty.
	if w.Grid[1][1].Type != CellEmpty {
		t.Errorf("expected empty at (1,1), got %d", w.Grid[1][1].Type)
	}
}

func TestSpawnPlayer(t *testing.T) {
	w := NewWorld()
	w.SpawnPlayer("p1", 0)
	w.SpawnPlayer("p2", 1)

	p1 := w.Players["p1"]
	if p1.X != 2 || p1.Y != 2 || p1.Dir != DirRight {
		t.Errorf("p1 spawn: got (%d,%d,%d), want (2,2,DirRight)", p1.X, p1.Y, p1.Dir)
	}
	p2 := w.Players["p2"]
	if p2.X != GridWidth-3 || p2.Y != GridHeight-3 || p2.Dir != DirLeft {
		t.Errorf("p2 spawn: got (%d,%d,%d), want (%d,%d,DirLeft)", p2.X, p2.Y, p2.Dir, GridWidth-3, GridHeight-3)
	}
}

func TestTakeSnapshot(t *testing.T) {
	w := NewWorld()
	w.SpawnPlayer("p1", 0)
	w.SpawnPlayer("p2", 1)

	snap := w.TakeSnapshot("p1")
	if len(snap.Players) != 2 {
		t.Errorf("expected 2 players in snapshot, got %d", len(snap.Players))
	}
	// Snapshot no longer computes game-over; that's done by tick().
	// Just verify it's a clean deep copy.
	if snap.Loser != "" || snap.Winner != "" {
		t.Error("expected no loser/winner on a fresh snapshot")
	}
}

func TestResetForWinner(t *testing.T) {
	w := NewWorld()
	w.SpawnPlayer("p1", 0)
	w.SpawnPlayer("p2", 1)

	// Lay some trail.
	w.Grid[5][5] = Cell{Type: CellTrail, Rune: 'x', OwnerID: "p1"}
	w.Players["p2"].Alive = false

	w.ResetForWinner("p1")

	// Grid interior should be empty.
	if w.Grid[5][5].Type != CellEmpty {
		t.Error("expected trail cleared after reset")
	}
	// Loser removed.
	if _, ok := w.Players["p2"]; ok {
		t.Error("expected p2 removed from world")
	}
	// Winner respawned at slot 0.
	p1 := w.Players["p1"]
	if p1.X != 2 || p1.Y != 2 || p1.Dir != DirRight {
		t.Errorf("winner respawn: got (%d,%d,%d), want (2,2,DirRight)", p1.X, p1.Y, p1.Dir)
	}
	if !p1.Alive {
		t.Error("winner should be alive after reset")
	}
}
