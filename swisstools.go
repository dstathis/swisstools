package swisstools

import (
	"errors"
	"fmt"
	"io"
	"math/rand"

	"github.com/olekukonko/tablewriter"
)

// Tournament constants
const (
	// Special player IDs and values
	BYE_OPPONENT_ID      = -1 // Player ID indicating a bye (no opponent)
	UNINITIALIZED_RESULT = -1 // Initial value for unset match results

	// Bye round scoring (tournament standard)
	BYE_WINS   = 2 // Games won when receiving a bye
	BYE_LOSSES = 0 // Games lost when receiving a bye
	BYE_DRAWS  = 0 // Games drawn when receiving a bye

	// Points system (tournament standard)
	POINTS_FOR_WIN  = 3 // Match points awarded for a win
	POINTS_FOR_DRAW = 1 // Match points awarded for a draw
	POINTS_FOR_LOSS = 0 // Match points awarded for a loss (explicit for clarity)
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
	player := Player{
		name:  name,
		notes: []string{},
		// points, wins, losses, draws are zero-initialized by Go
	}
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
			// Handle bye - last remaining player gets a bye
			t.rounds[t.currentRound] = append(t.rounds[t.currentRound],
				Pairing{playera: players[0], playerb: BYE_OPPONENT_ID, playeraWins: BYE_WINS, playerbWins: BYE_LOSSES, draws: BYE_DRAWS})
			break
		}

		// Pick two random players using helper function
		player0, remainingPlayers := removeRandomPlayer(players)
		player1, finalPlayers := removeRandomPlayer(remainingPlayers)
		players = finalPlayers

		// Create pairing between the two selected players
		t.rounds[t.currentRound] = append(t.rounds[t.currentRound],
			Pairing{playera: player0, playerb: player1, playeraWins: UNINITIALIZED_RESULT, playerbWins: UNINITIALIZED_RESULT, draws: UNINITIALIZED_RESULT})
	}
}

// removeRandomPlayer selects a random player from the slice and returns both
// the selected player and a new slice with that player removed.
func removeRandomPlayer(players []int) (int, []int) {
	if len(players) == 0 {
		panic("cannot remove player from empty slice")
	}

	// Pick random index
	index := rand.Intn(len(players))
	selectedPlayer := players[index]

	// Swap selected player with last element and shrink slice
	players[index] = players[len(players)-1]
	return selectedPlayer, players[:len(players)-1]
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
		// Check for incomplete matches (initialized with UNINITIALIZED_RESULT)
		if pairing.playeraWins == UNINITIALIZED_RESULT || pairing.playerbWins == UNINITIALIZED_RESULT || pairing.draws == UNINITIALIZED_RESULT {
			return errors.New("incomplete match found - all matches must have results")
		}
	}

	// SECOND PASS: All matches are complete, now update player stats
	for _, pairing := range t.rounds[t.currentRound] {
		// Handle bye rounds (playerb == BYE_OPPONENT_ID)
		if pairing.playerb == BYE_OPPONENT_ID {
			// Player gets a bye - worth POINTS_FOR_WIN (match win)
			playerA := t.players[pairing.playera]
			playerA.wins++
			playerA.points += POINTS_FOR_WIN
			t.players[pairing.playera] = playerA
			continue
		}

		// Determine match winner based on game results
		playerA := t.players[pairing.playera]
		playerB := t.players[pairing.playerb]

		if pairing.playeraWins > pairing.playerbWins {
			// Player A wins the match
			playerA.wins++
			playerA.points += POINTS_FOR_WIN
			playerB.losses++
			playerB.points += POINTS_FOR_LOSS // Explicit for clarity (currently 0)
		} else if pairing.playerbWins > pairing.playeraWins {
			// Player B wins the match
			playerB.wins++
			playerB.points += POINTS_FOR_WIN
			playerA.losses++
			playerA.points += POINTS_FOR_LOSS // Explicit for clarity (currently 0)
		} else {
			// Match is drawn (equal games won, or both 0 with draws > 0)
			playerA.draws++
			playerA.points += POINTS_FOR_DRAW
			playerB.draws++
			playerB.points += POINTS_FOR_DRAW
		}

		// Update players in the map
		t.players[pairing.playera] = playerA
		t.players[pairing.playerb] = playerB
	}

	return nil
}
