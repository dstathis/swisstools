package swisstools

import (
	"testing"
)

func TestAddPlayerName(t *testing.T) {
	tournament := NewTournament()
	err := tournament.AddPlayer("Dylan")
	if err != nil || tournament.players[1].name != "Dylan" {
		t.Fatalf("Expecting player name Dylan, got %s.", tournament.players[1].name)
	}
}

func TestAddPlayerEmpty(t *testing.T) {
	tournament := NewTournament()
	err := tournament.AddPlayer("")
	if err == nil {
		t.Fatalf("No name provided but AddPlayer did not return an error.")
	}
}

func TestAddResultBadId(t *testing.T) {
	tournament := NewTournament()
	err := tournament.AddResult(5, 2, 1, 0)
	if err == nil {
		t.Fatal("Bogus player submitted result but AddResult did not return an error.")
	}
}

func TestMultipleRounds(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")
	tournament.AddPlayer("Charlie")
	tournament.AddPlayer("Diana")

	// Test multiple rounds to ensure dynamic growth works
	for round := 1; round <= 5; round++ {
		tournament.Pair()
		if round < 5 {
			tournament.NextRound()
		}
	}
	t.Log("Successfully completed 5 rounds with dynamic slice growth.")
}

func TestPairRePairing(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")

	// First pairing should work
	tournament.Pair()
	round1 := tournament.GetRound()
	if len(round1) == 0 {
		t.Fatal("Expected pairings in round 1")
	}

	// Re-pairing the same round should work and clear previous pairings
	tournament.Pair()
	round1After := tournament.GetRound()
	if len(round1After) == 0 {
		t.Fatal("Expected pairings after re-pairing round 1")
	}

	// Should be able to pair after advancing to next round
	tournament.NextRound()
	tournament.Pair()
	round2 := tournament.GetRound()
	if len(round2) == 0 {
		t.Fatal("Expected pairings in round 2")
	}

	t.Log("Re-pairing functionality works correctly.")
}

func TestAddResultWithoutPairing(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")

	// Try to add result without calling Pair() first
	err := tournament.AddResult(1, 2, 0, 0)
	if err == nil {
		t.Fatal("Expected error when adding result without pairing, but got none")
	}
	if err.Error() != "round has no pairings - call Pair() first" {
		t.Fatalf("Expected specific error message, got: %v", err)
	}

	// Advance to next round and try again (NextRound creates empty round)
	tournament.NextRound()
	err = tournament.AddResult(1, 2, 0, 0)
	if err == nil {
		t.Fatal("Expected error when adding result to unpaired round, but got none")
	}
	if err.Error() != "round has no pairings - call Pair() first" {
		t.Fatalf("Expected specific error message, got: %v", err)
	}

	// After pairing, AddResult should work
	tournament.Pair()
	err = tournament.AddResult(1, 2, 0, 0)
	if err != nil {
		t.Fatalf("Expected AddResult to work after Pair(), got error: %v", err)
	}

	t.Log("Stronger defensive check works correctly.")
}
