package game

import "testing"

func TestParseCommand(t *testing.T) {
	tests := []struct {
		input string
		cmd   string
		valid bool
	}{
		{"up", "up", true},
		{"DOWN", "down", true},
		{" Left ", "left", true},
		{"right", "right", true},
		{"bomb", "bomb", true},
		{"foo", "", false},
		{"", "", false},
		{"UP  ", "up", true},
	}
	for _, tt := range tests {
		cmd, valid := ParseCommand(tt.input)
		if cmd != tt.cmd || valid != tt.valid {
			t.Errorf("ParseCommand(%q) = (%q, %v), want (%q, %v)", tt.input, cmd, valid, tt.cmd, tt.valid)
		}
	}
}
