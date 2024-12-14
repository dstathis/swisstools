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
