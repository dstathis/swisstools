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
