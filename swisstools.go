package swisstools

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"sort"

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
		// Byes must be handled separately because there's no opponent to update,
		// and the bye player automatically gets a match win with predetermined game scores
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

// Pair implements the proper Swiss tournament pairing algorithm.
func (t *Tournament) Pair(allowRepair bool) error {
	// Validate tournament state.
	if len(t.players) == 0 {
		return errors.New("cannot pair tournament with no players")
	}

	if t.currentRound < 1 {
		return errors.New("invalid tournament state: current round must be >= 1")
	}

	// Check if round already has pairings
	if t.currentRound < len(t.rounds) && len(t.rounds[t.currentRound]) > 0 {
		if !allowRepair {
			return errors.New("round already has pairings - use Pair(true) to allow re-pairing")
		}
		// Clear any existing pairings for this round to allow re-pairing
		t.rounds[t.currentRound] = Round{}
	}

	// Get players sorted by points (descending), with random ordering within same point groups
	players := t.getSortedPlayers()

	// Track which players have been paired
	paired := make(map[int]bool)
	var pairings []Pairing

	// First round: random pairing
	if t.currentRound == 1 {
		return t.randomPair()
	}

	// Subsequent rounds: Swiss pairing
	for i := 0; i < len(players); i++ {
		if paired[players[i]] {
			continue
		}

		// Find best available opponent
		opponent := t.findBestOpponent(players[i], players, paired)

		if opponent != -1 {
			// Create pairing
			pairings = append(pairings, Pairing{
				playera:     players[i],
				playerb:     opponent,
				playeraWins: UNINITIALIZED_RESULT,
				playerbWins: UNINITIALIZED_RESULT,
				draws:       UNINITIALIZED_RESULT,
			})
			paired[players[i]] = true
			paired[opponent] = true
		} else {
			// No opponent found, give bye
			pairings = append(pairings, Pairing{
				playera:     players[i],
				playerb:     BYE_OPPONENT_ID,
				playeraWins: BYE_WINS,
				playerbWins: BYE_LOSSES,
				draws:       BYE_DRAWS,
			})
			paired[players[i]] = true
		}
	}

	t.rounds[t.currentRound] = pairings
	return nil
}

// getSortedPlayers returns player IDs sorted by points (descending), with random ordering within same point groups
func (t *Tournament) getSortedPlayers() []int {
	var players []int
	for id := range t.players {
		players = append(players, id)
	}

	// Sort by points (descending) only
	sort.Slice(players, func(i, j int) bool {
		playerI := t.players[players[i]]
		playerJ := t.players[players[j]]
		return playerI.points > playerJ.points
	})

	// Randomize players within same point groups
	t.randomizeWithinPointGroups(players)

	return players
}

// randomizeWithinPointGroups randomizes the order of players within the same point groups
func (t *Tournament) randomizeWithinPointGroups(players []int) {
	if len(players) <= 1 {
		return
	}

	start := 0
	currentPoints := t.players[players[0]].points

	for i := 1; i < len(players); i++ {
		if t.players[players[i]].points != currentPoints {
			// Randomize the group from start to i-1
			if i-start > 1 {
				shufflePlayers(players[start:i])
			}
			start = i
			currentPoints = t.players[players[i]].points
		}
	}

	// Don't forget the last group
	if len(players)-start > 1 {
		shufflePlayers(players[start:])
	}
}

// shufflePlayers randomly shuffles a slice of player IDs
func shufflePlayers(players []int) {
	for i := len(players) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		players[i], players[j] = players[j], players[i]
	}
}

// findBestOpponent finds the best available opponent for a player
func (t *Tournament) findBestOpponent(playerID int, sortedPlayers []int, paired map[int]bool) int {
	player := t.players[playerID]

	// Look for opponents with same points first
	for _, opponentID := range sortedPlayers {
		if opponentID == playerID || paired[opponentID] {
			continue
		}

		if t.players[opponentID].points == player.points && !t.havePlayedBefore(playerID, opponentID) {
			return opponentID
		}
	}

	// If no same-point opponent, look for closest points
	for _, opponentID := range sortedPlayers {
		if opponentID == playerID || paired[opponentID] {
			continue
		}

		if !t.havePlayedBefore(playerID, opponentID) {
			return opponentID
		}
	}

	// If no opponent found without rematch, allow rematch as last resort
	for _, opponentID := range sortedPlayers {
		if opponentID == playerID || paired[opponentID] {
			continue
		}

		return opponentID
	}

	return -1 // No suitable opponent found
}

// havePlayedBefore checks if two players have played against each other in previous rounds
func (t *Tournament) havePlayedBefore(playerA, playerB int) bool {
	for round := 1; round < t.currentRound; round++ {
		if round >= len(t.rounds) {
			continue
		}

		for _, pairing := range t.rounds[round] {
			if (pairing.playera == playerA && pairing.playerb == playerB) ||
				(pairing.playera == playerB && pairing.playerb == playerA) {
				return true
			}
		}
	}
	return false
}

// randomPair implements the original random pairing logic
func (t *Tournament) randomPair() error {
	// Validate that we have players to pair
	if len(t.players) == 0 {
		return errors.New("cannot create random pairings with no players")
	}

	players := []int{}
	for id := range t.players {
		players = append(players, id)
	}

	var pairings []Pairing
	for len(players) > 0 {
		if len(players) == 1 {
			// Handle bye - last remaining player gets a bye
			pairings = append(pairings, Pairing{
				playera:     players[0],
				playerb:     BYE_OPPONENT_ID,
				playeraWins: BYE_WINS,
				playerbWins: BYE_LOSSES,
				draws:       BYE_DRAWS,
			})
			break
		}

		// Pick two random players using helper function
		player0, remainingPlayers := removeRandomPlayer(players)
		player1, finalPlayers := removeRandomPlayer(remainingPlayers)
		players = finalPlayers

		// Create pairing between the two selected players
		pairings = append(pairings, Pairing{
			playera:     player0,
			playerb:     player1,
			playeraWins: UNINITIALIZED_RESULT,
			playerbWins: UNINITIALIZED_RESULT,
			draws:       UNINITIALIZED_RESULT,
		})
	}

	t.rounds[t.currentRound] = pairings
	return nil
}
