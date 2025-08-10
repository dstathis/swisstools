package swisstools

import (
	"errors"
	"fmt"
	"io"
	"math/rand"

	"github.com/olekukonko/tablewriter"
)

type Tournament struct {
	lastId       int // Most recent player id to be assigned.
	players      map[int]Player
	currentRound int
	rounds       []Round
}

type Player struct {
	name   string
	points int
	wins   int
	losses int
	draws  int
	notes  []string
}

type Pairing struct {
	playera     int
	playerb     int
	playeraWins int
	playerbWins int
	draws       int
}

type Round = []Pairing

func NewTournament() Tournament {
	tournament := Tournament{}
	tournament.lastId = 0
	tournament.players = map[int]Player{}
	tournament.currentRound = 1          // Index round starting with 1 to make the round numbers human readable.
	tournament.rounds = make([]Round, 2) // Initialize with capacity for rounds 0 and 1
	return tournament
}

func (t *Tournament) AddPlayer(name string) error {
	if name == "" {
		return errors.New("empty name")
	}
	t.lastId++
	player := Player{}
	player.points = 0
	player.name = name
	player.notes = []string{}
	t.players[t.lastId] = player
	return nil
}

func (t *Tournament) FormatPlayers(w io.Writer) {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Name", "Wins", "Losses", "Points"})
	for _, player := range t.players {
		table.Append([]string{
			player.name,
			fmt.Sprintf("%d", player.wins),
			fmt.Sprintf("%d", player.losses),
			fmt.Sprintf("%d", player.points),
		})
	}
	table.Render()
}

func (t *Tournament) NextRound() error {
	err := t.UpdatePlayerStandings()
	if err != nil {
		return err
	}
	t.currentRound++
	// Ensure the rounds slice has capacity for the new round
	for len(t.rounds) <= t.currentRound {
		t.rounds = append(t.rounds, Round{})
	}
	return nil
}

func (t *Tournament) Pair() {
	// Clear any existing pairings for this round to allow re-pairing
	if t.currentRound < len(t.rounds) {
		t.rounds[t.currentRound] = Round{}
	}

	players := []int{}
	for id, _ := range t.players {
		players = append(players, id)
	}
	for len(players) > 0 {
		if len(players) == 1 {
			t.rounds[t.currentRound] = append(t.rounds[t.currentRound], Pairing{playera: players[0], playerb: -1, playeraWins: 2, playerbWins: 0, draws: 0})
			players = players[:0]
		} else {
			// Choose 2 random players and delete them from the list.
			playerIndex := rand.Intn(len(players))
			player0 := players[playerIndex]
			players[playerIndex] = players[len(players)-1]
			players = players[:len(players)-1]
			playerIndex = rand.Intn(len(players))
			player1 := players[playerIndex]
			players[playerIndex] = players[len(players)-1]
			players = players[:len(players)-1]
			t.rounds[t.currentRound] = append(t.rounds[t.currentRound], Pairing{playera: player0, playerb: player1, playeraWins: -1, playerbWins: -1, draws: -1})
		}
	}
}

func (t *Tournament) AddResult(id int, wins int, losses int, draws int) error {
	// Defensive check: ensure round exists and has been paired
	if t.currentRound >= len(t.rounds) {
		return errors.New("round not initialized - call NextRound() first")
	}
	if len(t.rounds[t.currentRound]) == 0 {
		return errors.New("round has no pairings - call Pair() first")
	}

	for i, pairing := range t.rounds[t.currentRound] {
		if pairing.playera == id {
			t.rounds[t.currentRound][i].playeraWins = wins
			t.rounds[t.currentRound][i].playerbWins = losses
			t.rounds[t.currentRound][i].draws = draws
			return nil
		}
		if pairing.playerb == id {
			t.rounds[t.currentRound][i].playerbWins = wins
			t.rounds[t.currentRound][i].playeraWins = losses
			t.rounds[t.currentRound][i].draws = draws
			return nil
		}
	}
	return errors.New("player not found")
}

func (t *Tournament) GetRound() []Pairing {
	// Defensive check - should not happen with proper NextRound() usage
	if t.currentRound >= len(t.rounds) {
		return []Pairing{} // Return empty slice if round not initialized
	}
	return t.rounds[t.currentRound]
}

// UpdatePlayerStandings processes the current round's pairings and updates player statistics.
// It calculates match wins/losses/draws and points based on game results within each pairing.
// Statistics are cumulative - this function adds to existing player stats.
// Returns an error if any matches in the current round are incomplete (have unset results).
// All matches must be complete before any player stats are updated (atomic operation).
func (t *Tournament) UpdatePlayerStandings() error {
	// Defensive check: ensure current round exists and has pairings
	if t.currentRound >= len(t.rounds) {
		return errors.New("round not initialized - call Pair() first")
	}
	if len(t.rounds[t.currentRound]) == 0 {
		return errors.New("round has no pairings - call Pair() first")
	}

	// FIRST PASS: Validate all matches are complete before updating any stats
	for _, pairing := range t.rounds[t.currentRound] {
		// Check for incomplete matches (initialized with -1)
		if pairing.playeraWins == -1 || pairing.playerbWins == -1 || pairing.draws == -1 {
			return errors.New("incomplete match found - all matches must have results")
		}
	}

	// SECOND PASS: All matches are complete, now update player stats
	for _, pairing := range t.rounds[t.currentRound] {
		// Handle bye rounds (playerb == -1)
		if pairing.playerb == -1 {
			// Player gets a bye - worth 3 points (match win)
			playerA := t.players[pairing.playera]
			playerA.wins++
			playerA.points += 3
			t.players[pairing.playera] = playerA
			continue
		}

		// Determine match winner based on game results
		playerA := t.players[pairing.playera]
		playerB := t.players[pairing.playerb]

		if pairing.playeraWins > pairing.playerbWins {
			// Player A wins the match
			playerA.wins++
			playerA.points += 3 // 3 points for a win
			playerB.losses++
		} else if pairing.playerbWins > pairing.playeraWins {
			// Player B wins the match
			playerB.wins++
			playerB.points += 3 // 3 points for a win
			playerA.losses++
		} else {
			// Match is drawn (equal games won, or both 0 with draws > 0)
			playerA.draws++
			playerA.points += 1 // 1 point for a draw
			playerB.draws++
			playerB.points += 1 // 1 point for a draw
		}

		// Update players in the map
		t.players[pairing.playera] = playerA
		t.players[pairing.playerb] = playerB
	}

	return nil
}
