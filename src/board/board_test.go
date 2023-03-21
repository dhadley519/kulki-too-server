package board

import "testing"

func TestAbs(t *testing.T) {
	got := abs(-1)
	if got != 1 {
		t.Errorf("Abs(-1) = %d; want 1", got)
	}
}

func TestBoard(t *testing.T) {
	b := NewBoard(9, 9, 6)
	//b.Set(&Position{1, 0}, 1)
	p, e := b.aStar(&Position{0, 0}, &Position{8, 8})
	if e != nil {
		t.Errorf("func returned error")
	}
	if len(p) != 17 {
		t.Errorf("unexpected number of moves: %d", len(p))
	}
}
