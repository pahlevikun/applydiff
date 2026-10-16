package application

import "testing"

func TestLegalTransition(t *testing.T) {
	to, err := Transition(StatusDiscovered, StatusMatched)
	if err != nil {
		t.Fatal(err)
	}
	if to != StatusMatched {
		t.Fatalf("got %s", to)
	}
}

func TestIllegalTransitionRejected(t *testing.T) {
	_, err := Transition(StatusDiscovered, StatusSubmitted)
	if err == nil {
		t.Fatal("expected error for discovered -> submitted")
	}
}

func TestClosedIsTerminal(t *testing.T) {
	if CanTransition(StatusClosed, StatusDiscovered) {
		t.Fatal("closed should not transition to discovered")
	}
}
