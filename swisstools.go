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

// TournamentConfig holds configuration options for tournaments
type TournamentConfig struct {
	PointsForWin  int // Points awarded for a win
	PointsForDraw int // Points awarded for a draw
	PointsForLoss int // Points awarded for a loss
	ByeWins       int // Games won when receiving a bye
	ByeLosses     int // Games lost when receiving a bye
	ByeDraws      int // Games drawn when receiving a bye
}

// DefaultConfig returns a default tournament configuration
func DefaultConfig() TournamentConfig {
	return TournamentConfig{
		PointsForWin:  POINTS_FOR_WIN,
		PointsForDraw: POINTS_FOR_DRAW,
		PointsForLoss: POINTS_FOR_LOSS,
		ByeWins:       BYE_WINS,
		ByeLosses:     BYE_LOSSES,
		ByeDraws:      BYE_DRAWS,
	}
}

type Tournament struct {
	config       TournamentConfig
	lastId       int // Most recent player id to be assigned.
	players      map[int]Player
	currentRound int
	rounds       []Round
	started      bool // Whether the tournament has started (first round paired)
	finished     bool // Whether the tournament has finished
}

// Decklist represents a player's deck configuration.
// Using maps allows counting unique items and future validation of quantities.
type Decklist struct {
	Main      map[string]int
	Sideboard map[string]int
}

type Player struct {
	name           string
	points         int
	wins           int
	losses         int
	draws          int
	gameWins       int
	gameLosses     int
	gameDraws      int
	notes          []string
	removed        bool // Whether the player has been removed from the tournament
	removedInRound int  // Round when the player was removed (0 if not removed)

	// Optional metadata
	externalID *int      // Optional global player identifier
	decklist   *Decklist // Optional decklist; validation not implemented yet
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
	return NewTournamentWithConfig(DefaultConfig())
}

func NewTournamentWithConfig(config TournamentConfig) Tournament {
	tournament := Tournament{}
	tournament.config = config
	tournament.lastId = 0
	tournament.players = map[int]Player{}
	tournament.currentRound = 1          // Index round starting with 1 to make the round numbers human readable.
	tournament.rounds = make([]Round, 2) // Initialize with capacity for rounds 0 and 1
	tournament.started = false
	tournament.finished = false
	return tournament
}

func (t *Tournament) AddPlayer(name string) error {
	if name == "" {
		return errors.New("empty name")
	}

	// Check if player with this name already exists
	for _, player := range t.players {
		if player.name == name {
			return errors.New("player with this name already exists")
		}
	}

	t.lastId++
	player := Player{
		name:  name,
		notes: []string{},
		// points, wins, losses, draws are zero-initialized by Go
	}
	t.players[t.lastId] = player

	// If tournament has started, add a note about late entry
	if t.started {
		player.notes = append(player.notes, fmt.Sprintf("Late entry - joined in round %d", t.currentRound))
		t.players[t.lastId] = player
	}

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

// StartTournament begins the tournament by creating the first round pairings
func (t *Tournament) StartTournament() error {
	if t.started {
		return errors.New("tournament has already started")
	}

	if len(t.players) == 0 {
		return errors.New("cannot start tournament with no players")
	}

	t.started = true
	return t.Pair(false)
}

// GetStatus returns the current tournament status
func (t *Tournament) GetStatus() string {
	if t.finished {
		return "finished"
	}
	if t.started {
		return "in_progress"
	}
	return "setup"
}

// GetCurrentRound returns the current round number
func (t *Tournament) GetCurrentRound() int {
	return t.currentRound
}

// GetPlayerCount returns the number of players in the tournament
func (t *Tournament) GetPlayerCount() int {
	return len(t.players)
}

// GetPlayerById returns a player by ID
func (t *Tournament) GetPlayerById(id int) (Player, bool) {
	player, exists := t.players[id]
	return player, exists
}

// GetPlayerID returns the ID of a player by name
func (t *Tournament) GetPlayerID(name string) (int, bool) {
	for id, player := range t.players {
		if player.name == name {
			return id, true
		}
	}
	return 0, false
}

// GetPlayerByName returns a player by name
func (t *Tournament) GetPlayerByName(name string) (Player, bool) {
	for _, player := range t.players {
		if player.name == name {
			return player, true
		}
	}
	return Player{}, false
}

// SetPlayerExternalID sets an optional global identifier for a player.
func (t *Tournament) SetPlayerExternalID(id int, externalID int) error {
	player, exists := t.players[id]
	if !exists {
		return errors.New("player not found")
	}
	player.externalID = &externalID
	t.players[id] = player
	return nil
}

// ClearPlayerExternalID removes the optional global identifier for a player.
func (t *Tournament) ClearPlayerExternalID(id int) error {
	player, exists := t.players[id]
	if !exists {
		return errors.New("player not found")
	}
	player.externalID = nil
	t.players[id] = player
	return nil
}

// GetPlayerExternalID returns the optional global identifier for a player.
// The boolean indicates if an identifier is set.
func (t *Tournament) GetPlayerExternalID(id int) (*int, bool) {
	player, exists := t.players[id]
	if !exists {
		return nil, false
	}
	if player.externalID == nil {
		return nil, false
	}
	return player.externalID, true
}

// SetPlayerDecklist sets the player's decklist.
func (t *Tournament) SetPlayerDecklist(id int, deck Decklist) error {
	player, exists := t.players[id]
	if !exists {
		return errors.New("player not found")
	}
	// Ensure maps are non-nil to avoid nil-map writes later
	if deck.Main == nil {
		deck.Main = map[string]int{}
	}
	if deck.Sideboard == nil {
		deck.Sideboard = map[string]int{}
	}
	dl := deck
	player.decklist = &dl
	t.players[id] = player
	return nil
}

// ClearPlayerDecklist removes the player's decklist.
func (t *Tournament) ClearPlayerDecklist(id int) error {
	player, exists := t.players[id]
	if !exists {
		return errors.New("player not found")
	}
	player.decklist = nil
	t.players[id] = player
	return nil
}

// GetPlayerDecklist returns the player's decklist if present.
func (t *Tournament) GetPlayerDecklist(id int) (*Decklist, bool) {
	player, exists := t.players[id]
	if !exists {
		return nil, false
	}
	if player.decklist == nil {
		return nil, false
	}
	return player.decklist, true
}

// RemovePlayerById removes a player from the tournament while preserving their history
func (t *Tournament) RemovePlayerById(id int) error {
	if _, exists := t.players[id]; !exists {
		return errors.New("player not found")
	}

	// If tournament has started, we need to handle the player's current round
	if t.started && t.currentRound < len(t.rounds) {
		// Find and handle any pairings involving this player in the current round
		for i, pairing := range t.rounds[t.currentRound] {
			if pairing.playera == id {
				if pairing.playerb == BYE_OPPONENT_ID {
					// Player has a bye, simply remove the pairing
					var newPairings []Pairing
					for j, p := range t.rounds[t.currentRound] {
						if j != i {
							newPairings = append(newPairings, p)
						}
					}
					t.rounds[t.currentRound] = newPairings
				} else {
					// Give opponent a bye
					t.rounds[t.currentRound][i] = Pairing{
						playera:     pairing.playerb,
						playerb:     BYE_OPPONENT_ID,
						playeraWins: t.config.ByeWins,
						playerbWins: t.config.ByeLosses,
						draws:       t.config.ByeDraws,
					}
				}
				break
			} else if pairing.playerb == id {
				if pairing.playera == BYE_OPPONENT_ID {
					// Player has a bye, simply remove the pairing
					var newPairings []Pairing
					for j, p := range t.rounds[t.currentRound] {
						if j != i {
							newPairings = append(newPairings, p)
						}
					}
					t.rounds[t.currentRound] = newPairings
				} else {
					// Give opponent a bye
					t.rounds[t.currentRound][i] = Pairing{
						playera:     pairing.playera,
						playerb:     BYE_OPPONENT_ID,
						playeraWins: t.config.ByeWins,
						playerbWins: t.config.ByeLosses,
						draws:       t.config.ByeDraws,
					}
				}
				break
			}
		}
	}

	// Mark player as removed (preserve history)
	player := t.players[id]
	player.removed = true
	player.removedInRound = t.currentRound
	t.players[id] = player

	return nil
}

// RemovePlayerByName removes a player from the tournament by name while preserving their history
func (t *Tournament) RemovePlayerByName(name string) error {
	// Find the player ID by name
	id, exists := t.GetPlayerID(name)
	if !exists {
		return errors.New("player not found")
	}

	// Use the existing RemovePlayerById method
	return t.RemovePlayerById(id)
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
			// Player gets a bye - worth PointsForWin (match win)
			playerA := t.players[pairing.playera]
			playerA.wins++
			playerA.points += t.config.PointsForWin
			// Add bye game results
			playerA.gameWins += t.config.ByeWins
			playerA.gameLosses += t.config.ByeLosses
			playerA.gameDraws += t.config.ByeDraws
			t.players[pairing.playera] = playerA
			continue
		}

		// Determine match winner based on game results
		playerA := t.players[pairing.playera]
		playerB := t.players[pairing.playerb]

		// Add individual game results
		playerA.gameWins += pairing.playeraWins
		playerA.gameLosses += pairing.playerbWins
		playerA.gameDraws += pairing.draws
		playerB.gameWins += pairing.playerbWins
		playerB.gameLosses += pairing.playeraWins
		playerB.gameDraws += pairing.draws

		if pairing.playeraWins > pairing.playerbWins {
			// Player A wins the match
			playerA.wins++
			playerA.points += t.config.PointsForWin
			playerB.losses++
			playerB.points += t.config.PointsForLoss // Explicit for clarity (currently 0)
		} else if pairing.playerbWins > pairing.playeraWins {
			// Player B wins the match
			playerB.wins++
			playerB.points += t.config.PointsForWin
			playerA.losses++
			playerA.points += t.config.PointsForLoss // Explicit for clarity (currently 0)
		} else {
			// Match is drawn (equal games won, or both 0 with draws > 0)
			playerA.draws++
			playerA.points += t.config.PointsForDraw
			playerB.draws++
			playerB.points += t.config.PointsForDraw
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
				playeraWins: t.config.ByeWins,
				playerbWins: t.config.ByeLosses,
				draws:       t.config.ByeDraws,
			})
			paired[players[i]] = true
		}
	}

	t.rounds[t.currentRound] = pairings
	return nil
}

// TiebreakerData holds calculated tiebreaker values for a player
type TiebreakerData struct {
	GameWinPercentage   float64 // Games won / total games played
	OpponentMatchWinPct float64 // Average match win percentage of opponents
	OpponentGameWinPct  float64 // Average game win percentage of opponents
}

// PlayerStanding represents a player's position in the tournament standings
type PlayerStanding struct {
	Rank        int
	PlayerID    int
	Name        string
	Points      int
	Wins        int
	Losses      int
	Draws       int
	Tiebreakers TiebreakerData
}

// calculateTiebreakers calculates all tiebreaker values for a player
func (t *Tournament) calculateTiebreakers(playerID int) TiebreakerData {
	player := t.players[playerID]

	// Calculate game win percentage
	totalGames := player.gameWins + player.gameLosses + player.gameDraws
	gameWinPct := 0.0
	if totalGames > 0 {
		gameWinPct = float64(player.gameWins) / float64(totalGames)
	}

	// Calculate opponent match win percentages
	opponentMatchWinPcts := []float64{}
	opponentGameWinPcts := []float64{}

	for round := 1; round < t.currentRound; round++ {
		if round >= len(t.rounds) {
			continue
		}

		for _, pairing := range t.rounds[round] {
			var opponentID int
			if pairing.playera == playerID {
				opponentID = pairing.playerb
			} else if pairing.playerb == playerID {
				opponentID = pairing.playera
			} else {
				continue
			}

			// Skip byes
			if opponentID == BYE_OPPONENT_ID {
				continue
			}

			// Get opponent data
			opponent, exists := t.players[opponentID]
			if !exists {
				continue
			}

			// Calculate opponent's match win percentage
			totalMatches := opponent.wins + opponent.losses + opponent.draws
			if totalMatches > 0 {
				opponentMatchWinPct := float64(opponent.wins) / float64(totalMatches)
				// Apply minimum 33% rule for Magic tournaments
				if opponentMatchWinPct < 0.33 {
					opponentMatchWinPct = 0.33
				}
				opponentMatchWinPcts = append(opponentMatchWinPcts, opponentMatchWinPct)
			}

			// Calculate opponent's game win percentage
			totalOpponentGames := opponent.gameWins + opponent.gameLosses + opponent.gameDraws
			if totalOpponentGames > 0 {
				opponentGameWinPct := float64(opponent.gameWins) / float64(totalOpponentGames)
				// Apply minimum 33% rule for Magic tournaments
				if opponentGameWinPct < 0.33 {
					opponentGameWinPct = 0.33
				}
				opponentGameWinPcts = append(opponentGameWinPcts, opponentGameWinPct)
			}
		}
	}

	// Calculate average opponent match win percentage
	avgOpponentMatchWinPct := 0.0
	if len(opponentMatchWinPcts) > 0 {
		sum := 0.0
		for _, pct := range opponentMatchWinPcts {
			sum += pct
		}
		avgOpponentMatchWinPct = sum / float64(len(opponentMatchWinPcts))
	}

	// Calculate average opponent game win percentage
	avgOpponentGameWinPct := 0.0
	if len(opponentGameWinPcts) > 0 {
		sum := 0.0
		for _, pct := range opponentGameWinPcts {
			sum += pct
		}
		avgOpponentGameWinPct = sum / float64(len(opponentGameWinPcts))
	}

	return TiebreakerData{
		GameWinPercentage:   gameWinPct,
		OpponentMatchWinPct: avgOpponentMatchWinPct,
		OpponentGameWinPct:  avgOpponentGameWinPct,
	}
}

// getSortedPlayersWithTiebreakers returns player IDs sorted by points and tiebreakers
func (t *Tournament) getSortedPlayersWithTiebreakers() []int {
	var players []int
	for id, player := range t.players {
		// Skip removed players
		if !player.removed {
			players = append(players, id)
		}
	}

	// Sort by points (descending), then by tiebreakers
	sort.Slice(players, func(i, j int) bool {
		playerI := t.players[players[i]]
		playerJ := t.players[players[j]]

		// First by points
		if playerI.points != playerJ.points {
			return playerI.points > playerJ.points
		}

		// Then by tiebreakers
		tiebreakersI := t.calculateTiebreakers(players[i])
		tiebreakersJ := t.calculateTiebreakers(players[j])

		// 1. Opponent match win percentage (first tiebreaker)
		if tiebreakersI.OpponentMatchWinPct != tiebreakersJ.OpponentMatchWinPct {
			return tiebreakersI.OpponentMatchWinPct > tiebreakersJ.OpponentMatchWinPct
		}

		// 2. Game win percentage (second tiebreaker)
		if tiebreakersI.GameWinPercentage != tiebreakersJ.GameWinPercentage {
			return tiebreakersI.GameWinPercentage > tiebreakersJ.GameWinPercentage
		}

		// 3. Opponent game win percentage (third tiebreaker)
		if tiebreakersI.OpponentGameWinPct != tiebreakersJ.OpponentGameWinPct {
			return tiebreakersI.OpponentGameWinPct > tiebreakersJ.OpponentGameWinPct
		}

		// If still tied, maintain original order (effectively random within same tiebreaker group)
		return i < j
	})

	return players
}

// getSortedPlayers returns player IDs sorted by points (descending), with random ordering within same point groups
func (t *Tournament) getSortedPlayers() []int {
	var players []int
	for id, player := range t.players {
		// Skip removed players
		if !player.removed {
			players = append(players, id)
		}
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
	for id, p := range t.players {
		if !p.removed {
			players = append(players, id)
		}
	}

	var pairings []Pairing
	for len(players) > 0 {
		if len(players) == 1 {
			// Handle bye - last remaining player gets a bye
			pairings = append(pairings, Pairing{
				playera:     players[0],
				playerb:     BYE_OPPONENT_ID,
				playeraWins: t.config.ByeWins,
				playerbWins: t.config.ByeLosses,
				draws:       t.config.ByeDraws,
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

// GetStandings returns the current tournament standings with tiebreakers
func (t *Tournament) GetStandings() []PlayerStanding {
	// Get players sorted by points and tiebreakers
	sortedPlayers := t.getSortedPlayersWithTiebreakers()

	var standings []PlayerStanding
	nextRank := 1

	for i, playerID := range sortedPlayers {
		player := t.players[playerID]
		tiebreakers := t.calculateTiebreakers(playerID)

		// Handle ties (same rank for players with identical points and tiebreakers)
		if i > 0 {
			prevPlayerID := sortedPlayers[i-1]
			prevPlayer := t.players[prevPlayerID]
			prevTiebreakers := t.calculateTiebreakers(prevPlayerID)

			if player.points == prevPlayer.points &&
				tiebreakers.OpponentMatchWinPct == prevTiebreakers.OpponentMatchWinPct &&
				tiebreakers.GameWinPercentage == prevTiebreakers.GameWinPercentage &&
				tiebreakers.OpponentGameWinPct == prevTiebreakers.OpponentGameWinPct {
				// Same rank as previous player (tied)
				nextRank = standings[len(standings)-1].Rank
			} else {
				// Count unique ranks used so far and add 1
				uniqueRanks := make(map[int]bool)
				for _, standing := range standings {
					uniqueRanks[standing.Rank] = true
				}
				nextRank = len(uniqueRanks) + 1
			}
		}

		standings = append(standings, PlayerStanding{
			Rank:        nextRank,
			PlayerID:    playerID,
			Name:        player.name,
			Points:      player.points,
			Wins:        player.wins,
			Losses:      player.losses,
			Draws:       player.draws,
			Tiebreakers: tiebreakers,
		})
	}

	return standings
}
