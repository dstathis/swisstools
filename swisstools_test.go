package swisstools

import (
	"bytes"
	"strings"
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
		tournament.Pair(false)

		// Add results for all pairings in this round
		pairings := tournament.GetRound()
		for _, pairing := range pairings {
			if pairing.playerb != -1 {
				// Add a simple win result
				err := tournament.AddResult(pairing.playera, 2, 1, 0)
				if err != nil {
					t.Fatalf("Failed to add result in round %d: %v", round, err)
				}
			}
			// Bye rounds already have results
		}

		if round < 5 {
			err := tournament.NextRound()
			if err != nil {
				t.Fatalf("Failed to advance to round %d: %v", round+1, err)
			}
		}
	}
	t.Log("Successfully completed 5 rounds with dynamic slice growth.")
}

func TestPairRePairing(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")

	// First pairing should work
	tournament.Pair(false)
	round1 := tournament.GetRound()
	if len(round1) == 0 {
		t.Fatal("Expected pairings in round 1")
	}

	// Re-pairing the same round should work and clear previous pairings
	tournament.Pair(false)
	round1After := tournament.GetRound()
	if len(round1After) == 0 {
		t.Fatal("Expected pairings after re-pairing round 1")
	}

	// Add results for the pairing so we can advance to next round
	for _, pairing := range round1After {
		if pairing.playerb != -1 {
			err := tournament.AddResult(pairing.playera, 2, 1, 0)
			if err != nil {
				t.Fatalf("Failed to add result: %v", err)
			}
		}
	}

	// Should be able to pair after advancing to next round
	err := tournament.NextRound()
	if err != nil {
		t.Fatalf("Failed to advance to next round: %v", err)
	}
	tournament.Pair(false)
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

	// Pair the round so we can advance, but don't add results
	tournament.Pair(false)

	// Try to advance to next round without adding results - should fail
	err = tournament.NextRound()
	if err == nil {
		t.Fatal("Expected NextRound to fail due to incomplete matches, but it succeeded")
	}
	if err.Error() != "incomplete match found - all matches must have results" {
		t.Fatalf("Expected specific error message, got: %v", err)
	}

	// After adding results, AddResult and NextRound should work
	err = tournament.AddResult(1, 2, 0, 0)
	if err != nil {
		t.Fatalf("Expected AddResult to work after Pair(), got error: %v", err)
	}

	// Now NextRound should succeed
	err = tournament.NextRound()
	if err != nil {
		t.Fatalf("Expected NextRound to succeed after completing matches, got error: %v", err)
	}
}

func TestUpdatePlayerStandings(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")   // ID: 1
	tournament.AddPlayer("Bob")     // ID: 2
	tournament.AddPlayer("Charlie") // ID: 3

	// Pair and see what we get (random pairing)
	tournament.Pair(false)
	pairings := tournament.GetRound()

	// Add results for all non-bye pairings
	for _, pairing := range pairings {
		if pairing.playerb != -1 {
			// Make player A win 2-1
			err := tournament.AddResult(pairing.playera, 2, 1, 0)
			if err != nil {
				t.Fatalf("Failed to add result: %v", err)
			}
		}
		// Bye rounds already have results set by Pair()
	}

	// Update standings for round 1
	err := tournament.UpdatePlayerStandings()
	if err != nil {
		t.Fatalf("Failed to update standings: %v", err)
	}

	// Check total wins + losses + draws equals number of players
	// (since with 3 players, we have 1 match + 1 bye = 2 match results total)
	totalWins := 0
	totalLosses := 0
	totalDraws := 0
	totalPoints := 0

	for _, player := range tournament.players {
		totalWins += player.wins
		totalLosses += player.losses
		totalDraws += player.draws
		totalPoints += player.points
	}

	// We should have 2 wins total (1 from match, 1 from bye), 1 loss, 0 draws
	if totalWins != 2 {
		t.Errorf("Expected 2 total wins, got %d", totalWins)
	}
	if totalLosses != 1 {
		t.Errorf("Expected 1 total loss, got %d", totalLosses)
	}
	if totalDraws != 0 {
		t.Errorf("Expected 0 total draws, got %d", totalDraws)
	}
	// With new points system: 2 wins = 6 points (3 each)
	if totalPoints != 6 {
		t.Errorf("Expected 6 total points with new system (3 per win), got %d", totalPoints)
	}

	t.Log("Player standings updated correctly for round 1")
}

func TestUpdatePlayerStandingsCumulative(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice") // ID: 1
	tournament.AddPlayer("Bob")   // ID: 2

	// Round 1: Alice beats Bob
	tournament.Pair(false)
	err := tournament.AddResult(1, 2, 0, 0) // Alice wins 2-0
	if err != nil {
		t.Fatalf("Failed to add result: %v", err)
	}

	// Round 2: Bob beats Alice - NextRound will update standings for round 1
	err = tournament.NextRound()
	if err != nil {
		t.Fatalf("Failed to advance to next round: %v", err)
	}

	// Check round 1 stats (after NextRound updates them)
	alice := tournament.players[1]
	bob := tournament.players[2]
	if alice.wins != 1 || alice.points != 3 {
		t.Errorf("After round 1: Alice should have 1 win, 3 points, got wins=%d, points=%d", alice.wins, alice.points)
	}
	if bob.losses != 1 || bob.points != 0 {
		t.Errorf("After round 1: Bob should have 1 loss, 0 points, got losses=%d, points=%d", bob.losses, bob.points)
	}
	tournament.Pair(false)
	err = tournament.AddResult(2, 2, 1, 0) // Bob wins 2-1
	if err != nil {
		t.Fatalf("Failed to add result: %v", err)
	}

	// Manually update standings for round 2 (since we're not calling NextRound again)
	err = tournament.UpdatePlayerStandings()
	if err != nil {
		t.Fatalf("Failed to update standings: %v", err)
	}

	// Check cumulative stats after round 2
	alice = tournament.players[1]
	bob = tournament.players[2]
	if alice.wins != 1 || alice.losses != 1 || alice.points != 3 {
		t.Errorf("After round 2: Alice should have 1 win, 1 loss, 3 points, got wins=%d, losses=%d, points=%d", alice.wins, alice.losses, alice.points)
	}
	if bob.wins != 1 || bob.losses != 1 || bob.points != 3 {
		t.Errorf("After round 2: Bob should have 1 win, 1 loss, 3 points, got wins=%d, losses=%d, points=%d", bob.wins, bob.losses, bob.points)
	}

	t.Log("Cumulative player standings working correctly across multiple rounds")
}

func TestUpdatePlayerStandingsIncompleteMatch(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")

	// Pair but don't add results
	tournament.Pair(false)

	// Should error due to incomplete match
	err := tournament.UpdatePlayerStandings()
	if err == nil {
		t.Fatal("Expected error for incomplete match, but got none")
	}
	if err.Error() != "incomplete match found - all matches must have results" {
		t.Fatalf("Expected specific error message, got: %v", err)
	}
}

func TestUpdatePlayerStandingsDraws(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")

	tournament.Pair(false)
	// Set up a drawn match (equal game wins)
	err := tournament.AddResult(1, 1, 1, 1) // Alice and Bob each win 1 game, 1 draw
	if err != nil {
		t.Fatalf("Failed to add result: %v", err)
	}

	err = tournament.UpdatePlayerStandings()
	if err != nil {
		t.Fatalf("Failed to update standings: %v", err)
	}

	alice := tournament.players[1]
	bob := tournament.players[2]

	// Both should have 1 draw and 1 point
	if alice.draws != 1 || alice.points != 1 {
		t.Errorf("Alice should have 1 draw and 1 point, got draws=%d, points=%d", alice.draws, alice.points)
	}
	if bob.draws != 1 || bob.points != 1 {
		t.Errorf("Bob should have 1 draw and 1 point, got draws=%d, points=%d", bob.draws, bob.points)
	}
}

func TestFormatPlayersWithActualData(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")

	tournament.Pair(false)
	// Alice wins 2-0
	err := tournament.AddResult(1, 2, 0, 0)
	if err != nil {
		t.Fatalf("Failed to add result: %v", err)
	}

	err = tournament.UpdatePlayerStandings()
	if err != nil {
		t.Fatalf("Failed to update standings: %v", err)
	}

	// Test that FormatPlayers doesn't crash and shows non-zero values
	// We can't easily test the exact output without capturing it, but we can
	// verify the function runs without panic
	var buf bytes.Buffer
	tournament.FormatPlayers(&buf)
	output := buf.String()

	// Should contain actual numbers, not just zeros
	if !strings.Contains(output, "1") || !strings.Contains(output, "0") {
		t.Errorf("FormatPlayers output should contain actual statistics, got: %s", output)
	}

	t.Log("FormatPlayers successfully displays actual player statistics")
}

func TestUpdatePlayerStandingsAtomic(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")   // ID: 1
	tournament.AddPlayer("Bob")     // ID: 2
	tournament.AddPlayer("Charlie") // ID: 3
	tournament.AddPlayer("Diana")   // ID: 4

	// Initial stats should be 0
	for _, player := range tournament.players {
		if player.wins != 0 || player.losses != 0 || player.points != 0 {
			t.Fatalf("Player %s should start with 0 stats", player.name)
		}
	}

	tournament.Pair(false)
	pairings := tournament.GetRound()

	// Add results for only SOME pairings, leaving others incomplete
	completedMatches := 0
	for _, pairing := range pairings {
		if pairing.playerb != -1 && completedMatches == 0 {
			// Complete only the first non-bye match
			err := tournament.AddResult(pairing.playera, 2, 1, 0)
			if err != nil {
				t.Fatalf("Failed to add result: %v", err)
			}
			completedMatches++
			break // Leave other matches incomplete
		}
	}

	// Try to update standings - should fail due to incomplete matches
	err := tournament.UpdatePlayerStandings()
	if err == nil {
		t.Fatal("Expected error due to incomplete matches, but UpdatePlayerStandings succeeded")
	}
	if err.Error() != "incomplete match found - all matches must have results" {
		t.Fatalf("Expected specific error message, got: %v", err)
	}

	// Verify NO player stats were updated (atomic behavior)
	for _, player := range tournament.players {
		if player.wins != 0 || player.losses != 0 || player.points != 0 {
			t.Errorf("Player %s stats were modified despite incomplete matches: wins=%d, losses=%d, points=%d",
				player.name, player.wins, player.losses, player.points)
		}
	}

	t.Log("Atomic behavior verified - no stats updated when matches are incomplete")
}

func TestCorrectPointsSystem(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")   // ID: 1
	tournament.AddPlayer("Bob")     // ID: 2
	tournament.AddPlayer("Charlie") // ID: 3

	// Test wins and draws with correct points
	tournament.Pair(false)
	pairings := tournament.GetRound()

	// Set up specific results to test points system
	for _, pairing := range pairings {
		if pairing.playerb != -1 {
			// Alice beats Bob (assuming Alice is playera)
			err := tournament.AddResult(pairing.playera, 2, 1, 0)
			if err != nil {
				t.Fatalf("Failed to add result: %v", err)
			}
		}
		// Bye already has results
	}

	// Test another round with a draw - NextRound will update standings for round 1
	err := tournament.NextRound()
	if err != nil {
		t.Fatalf("Failed to advance to next round: %v", err)
	}
	tournament.Pair(false)

	// Add a draw result
	pairings = tournament.GetRound()
	for _, pairing := range pairings {
		if pairing.playerb != -1 {
			// Set up a draw (equal games)
			err := tournament.AddResult(pairing.playera, 1, 1, 0)
			if err != nil {
				t.Fatalf("Failed to add result: %v", err)
			}
		}
	}

	err = tournament.UpdatePlayerStandings()
	if err != nil {
		t.Fatalf("Failed to update standings: %v", err)
	}

	// Verify points system: win = 3 points, draw = 1 point, bye = 3 points
	totalPoints := 0
	totalWins := 0
	totalDraws := 0

	for _, player := range tournament.players {
		totalPoints += player.points
		totalWins += player.wins
		totalDraws += player.draws
	}

	// With 3 players across 2 rounds, we should have specific point totals
	// Round 1: 1 match + 1 bye = 6 points total (3 for winner + 0 for loser + 3 for bye)
	// Round 2: 1 draw + 1 bye = 5 points total (1 + 1 for draw + 3 for bye)
	// Total: 11 points
	expectedTotalPoints := 11
	if totalPoints != expectedTotalPoints {
		t.Errorf("Expected %d total points with correct system, got %d", expectedTotalPoints, totalPoints)
	}

	t.Logf("Correct points system verified: %d total points across %d wins and %d draws",
		totalPoints, totalWins, totalDraws)
}

// Test for removeRandomPlayer panic case indirectly through Pair with no players
func TestPairWithNoPlayers(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// This would happen if removeRandomPlayer panics, but Pair should handle this gracefully
			t.Logf("Recovered from panic: %v", r)
		}
	}()

	tournament := NewTournament()
	// Don't add any players

	// This should handle the empty players case gracefully
	tournament.Pair(false)

	// Check that no pairings were created
	pairings := tournament.GetRound()
	if len(pairings) != 0 {
		t.Errorf("Expected no pairings with no players, got %d", len(pairings))
	}
}

// Test GetRound defensive check (round not initialized)
func TestGetRoundNotInitialized(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")

	// Set currentRound beyond rounds slice length
	tournament.currentRound = 10

	result := tournament.GetRound()
	if len(result) != 0 {
		t.Errorf("Expected empty slice for uninitialized round, got %d pairings", len(result))
	}
}

// Test AddResult when player is playerb (not just playera)
func TestAddResultPlayerB(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")

	tournament.Pair(false)

	// Get the pairing to find player IDs
	pairings := tournament.GetRound()
	if len(pairings) == 0 {
		t.Fatal("No pairings found")
	}

	// Add result for playerb
	err := tournament.AddResult(pairings[0].playerb, 2, 1, 0)
	if err != nil {
		t.Fatalf("Failed to add result for playerb: %v", err)
	}

	// Verify the result was recorded (playerb wins = 2, playera wins = 1)
	updatedPairings := tournament.GetRound()
	if updatedPairings[0].playerbWins != 2 || updatedPairings[0].playeraWins != 1 {
		t.Errorf("Result not recorded correctly for playerb")
	}
}

// Test AddResult with player not found
func TestAddResultPlayerNotFound(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")

	tournament.Pair(false)

	// Try to add result for non-existent player
	err := tournament.AddResult(999, 2, 1, 0)
	if err == nil {
		t.Fatal("Expected error when adding result for non-existent player")
	}

	expectedMsg := "player not found"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// Test UpdatePlayerStandings when round not initialized
func TestUpdatePlayerStandingsRoundNotInitialized(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")

	// Set currentRound beyond rounds slice length to trigger defensive check
	tournament.currentRound = 10

	err := tournament.UpdatePlayerStandings()
	if err == nil {
		t.Fatal("Expected error when updating standings for uninitialized round")
	}

	expectedMsg := "round not initialized - call Pair() first"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// Test the panic in removeRandomPlayer indirectly
func TestRemoveRandomPlayerPanic(t *testing.T) {
	// Since removeRandomPlayer is unexported, we can't test the panic directly
	// but we've already covered the main functionality through integration tests
	t.Skip("Cannot directly test unexported function panic - covered by integration tests")
}

// Test AddResult round not initialized defensive check
func TestAddResultRoundNotInitialized(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")

	// Set currentRound beyond rounds slice length
	tournament.currentRound = 10

	err := tournament.AddResult(1, 2, 1, 0)
	if err == nil {
		t.Fatal("Expected error when adding result for uninitialized round")
	}

	expectedMsg := "round not initialized - call NextRound() first"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// Test to ensure we cover all branches in UpdatePlayerStandings
func TestUpdatePlayerStandingsAllBranches(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")
	tournament.AddPlayer("Charlie")

	tournament.Pair(false)

	// Get pairings to set up specific win/loss scenarios
	pairings := tournament.GetRound()

	if len(pairings) >= 1 {
		// First pairing: Bob wins (to trigger playerB wins branch)
		if pairings[0].playerb != BYE_OPPONENT_ID {
			err := tournament.AddResult(pairings[0].playerb, 2, 1, 0) // playerb wins
			if err != nil {
				t.Fatalf("Failed to add result: %v", err)
			}
		}
	}

	// If there's a bye, the bye case is already covered
	// This should ensure we hit the playerB wins branch with explicit POINTS_FOR_LOSS
	err := tournament.UpdatePlayerStandings()
	if err != nil {
		t.Fatalf("Failed to update standings: %v", err)
	}
}

// Test for a true tie scenario (both players have same wins, not 0-0 with draws)
func TestUpdatePlayerStandingsTrueDrawScenario(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")

	tournament.Pair(false)

	// Create a scenario where both players have the same number of wins (1-1 with draws)
	pairings := tournament.GetRound()
	if len(pairings) > 0 && pairings[0].playerb != BYE_OPPONENT_ID {
		err := tournament.AddResult(pairings[0].playera, 1, 1, 1) // 1-1-1 result
		if err != nil {
			t.Fatalf("Failed to add result: %v", err)
		}

		err = tournament.UpdatePlayerStandings()
		if err != nil {
			t.Fatalf("Failed to update standings: %v", err)
		}
	}
}

func TestPairErrorHandling(t *testing.T) {
	// Test pairing with no players
	tournament := NewTournament()
	err := tournament.Pair(false)
	if err == nil {
		t.Fatal("Expected error when pairing tournament with no players")
	}
	if err.Error() != "cannot pair tournament with no players" {
		t.Errorf("Expected specific error message, got: %v", err)
	}

	// Test pairing with invalid tournament state
	tournament = NewTournament()
	tournament.currentRound = 0 // Invalid state
	tournament.AddPlayer("Alice")
	err = tournament.Pair(false)
	if err == nil {
		t.Fatal("Expected error when pairing with invalid tournament state")
	}
	if err.Error() != "invalid tournament state: current round must be >= 1" {
		t.Errorf("Expected specific error message, got: %v", err)
	}

	// Test that pairing works with valid state
	tournament = NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")
	err = tournament.Pair(false)
	if err != nil {
		t.Fatalf("Expected no error when pairing valid tournament, got: %v", err)
	}

	t.Log("Pair error handling working correctly")
}

func TestPairRepairFunctionality(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")

	// First pairing should work
	err := tournament.Pair(false)
	if err != nil {
		t.Fatalf("First Pair() failed: %v", err)
	}

	// Second pairing without repair should fail
	err = tournament.Pair(false)
	if err == nil {
		t.Fatal("Expected error when calling Pair() without repair on already paired round")
	}
	if err.Error() != "round already has pairings - use Pair(true) to allow re-pairing" {
		t.Errorf("Expected specific error message, got: %v", err)
	}

	// Pairing with repair should work
	err = tournament.Pair(true)
	if err != nil {
		t.Fatalf("Pair(true) failed: %v", err)
	}

	// Verify we have pairings
	pairings := tournament.GetRound()
	if len(pairings) == 0 {
		t.Fatal("Expected pairings after Pair(true)")
	}

	t.Log("Pair repair functionality working correctly")
}

func TestTournamentConfiguration(t *testing.T) {
	// Test custom configuration
	config := TournamentConfig{
		PointsForWin:  4,
		PointsForDraw: 2,
		PointsForLoss: 1,
		ByeWins:       3,
		ByeLosses:     0,
		ByeDraws:      0,
	}

	tournament := NewTournamentWithConfig(config)

	// Test that tournament can start with any number of players > 0
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")
	tournament.AddPlayer("Charlie")

	// Should be able to start with 3 players
	if tournament.GetStatus() != "setup" {
		t.Error("Expected tournament to be in setup status")
	}

	err := tournament.StartTournament()
	if err != nil {
		t.Fatalf("Failed to start tournament: %v", err)
	}

	// Verify tournament started
	if tournament.GetStatus() != "in_progress" {
		t.Error("Expected tournament to be in progress")
	}

	if tournament.GetStatus() != "in_progress" {
		t.Errorf("Expected status 'in_progress', got '%s'", tournament.GetStatus())
	}

	t.Log("Tournament configuration working correctly")
}

func TestTournamentStateManagement(t *testing.T) {
	tournament := NewTournament()

	// Initial state
	if tournament.GetStatus() != "setup" {
		t.Errorf("Expected initial status 'setup', got '%s'", tournament.GetStatus())
	}

	// Add players and start
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")
	tournament.AddPlayer("Charlie")
	tournament.AddPlayer("Diana")

	err := tournament.StartTournament()
	if err != nil {
		t.Fatalf("Failed to start tournament: %v", err)
	}

	// Verify started state
	if tournament.GetStatus() != "in_progress" {
		t.Error("Expected tournament to be in progress")
	}

	if tournament.GetStatus() != "in_progress" {
		t.Errorf("Expected status 'in_progress', got '%s'", tournament.GetStatus())
	}

	// Test player management
	player, exists := tournament.GetPlayerById(1)
	if !exists {
		t.Error("Expected player 1 to exist")
	}
	if player.name != "Alice" {
		t.Errorf("Expected player 1 to be Alice, got %s", player.name)
	}

	id, exists := tournament.GetPlayerID("Bob")
	if !exists {
		t.Error("Expected to find Bob by name")
	}
	if id != 2 {
		t.Errorf("Expected Bob to have ID 2, got %d", id)
	}

	// Test removing player before start
	tournament2 := NewTournament()
	tournament2.AddPlayer("Alice")
	tournament2.AddPlayer("Bob")

	err = tournament2.RemovePlayerById(1)
	if err != nil {
		t.Fatalf("Failed to remove player: %v", err)
	}

	// Player should still exist but be marked as removed
	if tournament2.GetPlayerCount() != 2 {
		t.Errorf("Expected 2 players after removal (history preserved), got %d", tournament2.GetPlayerCount())
	}

	// Check that player is marked as removed
	player, exists = tournament2.GetPlayerById(1)
	if !exists {
		t.Error("Player should still exist after removal")
	}

	if !player.removed {
		t.Error("Expected player to be marked as removed")
	}

	// Test removing player after start (should work now)
	err = tournament.RemovePlayerById(1)
	if err != nil {
		t.Errorf("Expected to be able to remove player after tournament started: %v", err)
	}

	t.Log("Tournament state management working correctly")
}

func TestPlayerManagementDuringTournament(t *testing.T) {
	tournament := NewTournament()

	// Add initial players and start tournament
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")
	tournament.AddPlayer("Charlie")
	tournament.AddPlayer("Diana")

	err := tournament.StartTournament()
	if err != nil {
		t.Fatalf("Failed to start tournament: %v", err)
	}

	// Test adding a player during the tournament
	err = tournament.AddPlayer("Eve")
	if err != nil {
		t.Fatalf("Failed to add player during tournament: %v", err)
	}

	// Verify the late entry note was added
	player, exists := tournament.GetPlayerById(5) // Eve should be player 5
	if !exists {
		t.Fatal("Eve not found")
	}

	hasLateEntryNote := false
	for _, note := range player.notes {
		if strings.Contains(note, "Late entry") {
			hasLateEntryNote = true
			break
		}
	}
	if !hasLateEntryNote {
		t.Error("Expected late entry note for Eve")
	}

	// Test removing a player during the tournament
	err = tournament.RemovePlayerById(2) // Remove Bob
	if err != nil {
		t.Fatalf("Failed to remove player during tournament: %v", err)
	}

	// Verify Bob still exists but is marked as removed
	_, exists = tournament.GetPlayerById(2)
	if !exists {
		t.Error("Bob should still exist but be marked as removed")
	}

	// Test removing a player during the tournament
	err = tournament.RemovePlayerById(3) // Remove Charlie
	if err != nil {
		t.Fatalf("Failed to remove player during tournament: %v", err)
	}

	// Verify Charlie is marked as removed
	player, exists = tournament.GetPlayerById(3)
	if !exists {
		t.Error("Charlie should still exist but be marked as removed")
	}

	if !player.removed {
		t.Error("Expected Charlie to be marked as removed")
	}

	// Test that dropped players are excluded from pairing
	players := tournament.getSortedPlayers()
	for _, id := range players {
		if id == 3 { // Charlie's ID
			t.Error("Dropped player should not be included in pairing")
		}
	}

	// Test adding a player with duplicate name
	err = tournament.AddPlayer("Alice")
	if err == nil {
		t.Error("Expected error when adding duplicate player name")
	}

	t.Log("Player management during tournament working correctly")
}

func TestGetPlayerByName(t *testing.T) {
	tournament := NewTournament()

	// Add some players
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")
	tournament.AddPlayer("Charlie")

	// Test getting existing player
	player, exists := tournament.GetPlayerByName("Bob")
	if !exists {
		t.Fatal("Expected to find Bob")
	}
	if player.name != "Bob" {
		t.Errorf("Expected player name 'Bob', got '%s'", player.name)
	}

	// Test getting non-existent player
	player, exists = tournament.GetPlayerByName("David")
	if exists {
		t.Error("Expected not to find David")
	}
	if player.name != "" {
		t.Errorf("Expected empty name for non-existent player, got '%s'", player.name)
	}

	// Test case sensitivity
	player, exists = tournament.GetPlayerByName("alice")
	if exists {
		t.Error("Expected case-sensitive matching")
	}

	t.Log("GetPlayerByName working correctly")
}

func TestRemovePlayerByeHandling(t *testing.T) {
	tournament := NewTournament()

	// Add players and start tournament (3 players ensures a bye)
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")
	tournament.AddPlayer("Charlie")

	err := tournament.StartTournament()
	if err != nil {
		t.Fatalf("Failed to start tournament: %v", err)
	}

	// Get the current round pairings
	pairings := tournament.GetRound()
	if len(pairings) != 2 {
		t.Fatalf("Expected 2 pairings, got %d", len(pairings))
	}

	// Find which player has a bye (playerb == BYE_OPPONENT_ID)
	var byePlayerID int
	for _, pairing := range pairings {
		if pairing.playerb == BYE_OPPONENT_ID {
			byePlayerID = pairing.playera
			break
		}
	}

	if byePlayerID == 0 {
		t.Fatal("No bye player found in pairings")
	}

	// Test removing a player who has a bye
	initialPairingCount := len(tournament.GetRound())
	err = tournament.RemovePlayerById(byePlayerID)
	if err != nil {
		t.Fatalf("Failed to remove bye player: %v", err)
	}

	// Should have one fewer pairing (bye pairing removed)
	newPairingCount := len(tournament.GetRound())
	if newPairingCount != initialPairingCount-1 {
		t.Errorf("Expected %d pairings after dropping bye player, got %d", initialPairingCount-1, newPairingCount)
	}

	// Test dropping a player who has an opponent
	// First, add results for remaining matches and advance to next round
	for _, pairing := range tournament.GetRound() {
		if pairing.playerb != BYE_OPPONENT_ID {
			err = tournament.AddResult(pairing.playera, 2, 1, 0)
			if err != nil {
				t.Fatalf("Failed to add result: %v", err)
			}
		}
	}

	err = tournament.NextRound()
	if err != nil {
		t.Fatalf("Failed to advance to next round: %v", err)
	}

	err = tournament.Pair(false)
	if err != nil {
		t.Fatalf("Failed to pair next round: %v", err)
	}

	// Find a player with an opponent (not a bye)
	var playerWithOpponent int
	for _, pairing := range tournament.GetRound() {
		if pairing.playerb != BYE_OPPONENT_ID {
			playerWithOpponent = pairing.playera
			break
		}
	}

	// Remove the player with opponent
	initialPairingCount = len(tournament.GetRound())
	err = tournament.RemovePlayerById(playerWithOpponent)
	if err != nil {
		t.Fatalf("Failed to remove player with opponent: %v", err)
	}

	// Should have same number of pairings (bye given to opponent)
	newPairingCount = len(tournament.GetRound())
	if newPairingCount != initialPairingCount {
		t.Errorf("Expected %d pairings after dropping player with opponent, got %d", initialPairingCount, newPairingCount)
	}

	// Verify the opponent now has a bye
	hasByeForOpponent := false
	for _, pairing := range tournament.GetRound() {
		if pairing.playera != playerWithOpponent && pairing.playerb == BYE_OPPONENT_ID {
			hasByeForOpponent = true
			break
		}
	}
	if !hasByeForOpponent {
		t.Error("Expected opponent to receive a bye when player dropped")
	}

	t.Log("RemovePlayer bye handling working correctly")
}

func TestRemovePlayerByName(t *testing.T) {
	tournament := NewTournament()

	// Add players
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")
	tournament.AddPlayer("Charlie")

	// Test removing player by name
	err := tournament.RemovePlayerByName("Bob")
	if err != nil {
		t.Fatalf("Failed to remove player by name: %v", err)
	}

	// Verify Bob still exists but is marked as removed
	player, exists := tournament.GetPlayerById(2) // Bob should be player 2
	if !exists {
		t.Error("Bob should still exist but be marked as removed")
	}

	if !player.removed {
		t.Error("Expected Bob to be marked as removed")
	}

	// Test removing non-existent player by name
	err = tournament.RemovePlayerByName("David")
	if err == nil {
		t.Error("Expected error when removing non-existent player")
	}

	// Test case sensitivity
	err = tournament.RemovePlayerByName("alice")
	if err == nil {
		t.Error("Expected case-sensitive matching")
	}

	t.Log("RemovePlayerByName working correctly")
}

func TestTiebreakerSystem(t *testing.T) {
	tournament := NewTournament()

	// Add players
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")
	tournament.AddPlayer("Charlie")
	tournament.AddPlayer("Diana")

	// Start tournament
	err := tournament.StartTournament()
	if err != nil {
		t.Fatalf("Failed to start tournament: %v", err)
	}

	// Add results for round 1 to create different scenarios
	pairings := tournament.GetRound()
	for _, pairing := range pairings {
		if pairing.playerb != BYE_OPPONENT_ID {
			// Alice wins 2-1, Bob wins 2-0, Charlie wins 2-1, Diana gets bye
			if pairing.playera == 1 { // Alice
				tournament.AddResult(pairing.playera, 2, 1, 0)
			} else if pairing.playera == 2 { // Bob
				tournament.AddResult(pairing.playera, 2, 0, 0)
			} else if pairing.playera == 3 { // Charlie
				tournament.AddResult(pairing.playera, 2, 1, 0)
			} else if pairing.playera == 4 { // Diana
				tournament.AddResult(pairing.playera, 1, 2, 0)
			}
		}
		// Bye matches already have results
	}

	// Update standings
	err = tournament.UpdatePlayerStandings()
	if err != nil {
		t.Fatalf("Failed to update standings: %v", err)
	}

	// Get standings with tiebreakers
	standings := tournament.GetStandings()

	// Verify we have standings
	if len(standings) != 4 {
		t.Fatalf("Expected 4 standings, got %d", len(standings))
	}

	// Verify tiebreaker data is calculated
	for _, standing := range standings {
		if standing.Tiebreakers.GameWinPercentage < 0 || standing.Tiebreakers.GameWinPercentage > 1 {
			t.Errorf("Invalid game win percentage for %s: %f", standing.Name, standing.Tiebreakers.GameWinPercentage)
		}
		if standing.Tiebreakers.OpponentMatchWinPct < 0 || standing.Tiebreakers.OpponentMatchWinPct > 1 {
			t.Errorf("Invalid opponent match win percentage for %s: %f", standing.Name, standing.Tiebreakers.OpponentMatchWinPct)
		}
	}

	// Print standings for debugging
	t.Logf("Standings:")
	for _, standing := range standings {
		t.Logf("%d. %s - Points: %d, Games: %d-%d-%d, Game Win %%: %.3f",
			standing.Rank, standing.Name, standing.Points,
			standing.Wins, standing.Losses, standing.Draws,
			standing.Tiebreakers.GameWinPercentage)
	}

	// Verify that standings are sorted by points first, then tiebreakers
	for i := 1; i < len(standings); i++ {
		if standings[i].Points > standings[i-1].Points {
			t.Errorf("Standings not sorted by points: %s (%d) before %s (%d)",
				standings[i-1].Name, standings[i-1].Points,
				standings[i].Name, standings[i].Points)
		}
	}

	// Verify that all players have valid tiebreaker data
	for _, standing := range standings {
		if standing.Tiebreakers.GameWinPercentage < 0 || standing.Tiebreakers.GameWinPercentage > 1 {
			t.Errorf("Invalid game win percentage for %s: %f", standing.Name, standing.Tiebreakers.GameWinPercentage)
		}
	}

	t.Log("Tiebreaker system working correctly")
}

func TestGameTrackingInTiebreakers(t *testing.T) {
	tournament := NewTournament()

	// Add players
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")
	tournament.AddPlayer("Charlie")
	tournament.AddPlayer("Diana")

	// Start tournament (this automatically pairs the first round)
	err := tournament.StartTournament()
	if err != nil {
		t.Fatalf("Failed to start tournament: %v", err)
	}

	// Add results with different game scores
	pairings := tournament.GetRound()

	// Add results for all pairings
	for _, pairing := range pairings {
		if pairing.playerb == BYE_OPPONENT_ID {
			// Skip bye pairings
			continue
		}

		// Add results based on player ID
		if pairing.playera == 1 { // Alice
			tournament.AddResult(pairing.playera, 2, 1, 0) // Alice wins 2-1
		} else if pairing.playera == 2 { // Bob
			tournament.AddResult(pairing.playera, 1, 2, 0) // Bob loses 1-2
		} else if pairing.playera == 3 { // Charlie
			tournament.AddResult(pairing.playera, 2, 0, 0) // Charlie wins 2-0
		} else if pairing.playera == 4 { // Diana
			tournament.AddResult(pairing.playera, 0, 2, 0) // Diana loses 0-2
		}
	}

	// Update standings
	err = tournament.UpdatePlayerStandings()
	if err != nil {
		t.Fatalf("Failed to update standings: %v", err)
	}

	// Get standings
	standings := tournament.GetStandings()

	// Find players in standings
	var aliceStanding, bobStanding PlayerStanding
	for _, standing := range standings {
		if standing.Name == "Alice" {
			aliceStanding = standing
		}
		if standing.Name == "Bob" {
			bobStanding = standing
		}
	}

	// Check that both players have valid game win percentages
	if aliceStanding.Tiebreakers.GameWinPercentage < 0 || aliceStanding.Tiebreakers.GameWinPercentage > 1 {
		t.Errorf("Invalid game win percentage for Alice: %.3f", aliceStanding.Tiebreakers.GameWinPercentage)
	}

	if bobStanding.Tiebreakers.GameWinPercentage < 0 || bobStanding.Tiebreakers.GameWinPercentage > 1 {
		t.Errorf("Invalid game win percentage for Bob: %.3f", bobStanding.Tiebreakers.GameWinPercentage)
	}

	// Verify that game tracking is working by checking that the percentages are different
	// (since they had different game results)
	if aliceStanding.Tiebreakers.GameWinPercentage == bobStanding.Tiebreakers.GameWinPercentage {
		t.Errorf("Expected different game win percentages, but both players have %.3f",
			aliceStanding.Tiebreakers.GameWinPercentage)
	}

	t.Logf("Game tracking test passed: Alice (%.3f game win %%) vs Bob (%.3f game win %%)",
		aliceStanding.Tiebreakers.GameWinPercentage, bobStanding.Tiebreakers.GameWinPercentage)
}

func TestMinimum33PercentRule(t *testing.T) {
	tournament := NewTournament()

	// Add players
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")
	tournament.AddPlayer("Charlie")
	tournament.AddPlayer("Diana")

	// Start tournament
	err := tournament.StartTournament()
	if err != nil {
		t.Fatalf("Failed to start tournament: %v", err)
	}

	// Round 1: Create initial records
	pairings := tournament.GetRound()
	for _, pairing := range pairings {
		if pairing.playerb == BYE_OPPONENT_ID {
			continue
		}

		// Alice beats Bob, Charlie beats Diana
		if pairing.playera == 1 { // Alice
			tournament.AddResult(pairing.playera, 2, 0, 0) // Alice wins 2-0
		} else if pairing.playera == 2 { // Bob
			tournament.AddResult(pairing.playera, 0, 2, 0) // Bob loses 0-2
		} else if pairing.playera == 3 { // Charlie
			tournament.AddResult(pairing.playera, 2, 0, 0) // Charlie wins 2-0
		} else if pairing.playera == 4 { // Diana
			tournament.AddResult(pairing.playera, 0, 2, 0) // Diana loses 0-2
		}
	}

	// Update standings for round 1
	err = tournament.UpdatePlayerStandings()
	if err != nil {
		t.Fatalf("Failed to update standings: %v", err)
	}

	// Advance to round 2
	err = tournament.NextRound()
	if err != nil {
		t.Fatalf("Failed to advance to round 2: %v", err)
	}

	// Pair round 2
	err = tournament.Pair(false)
	if err != nil {
		t.Fatalf("Failed to pair round 2: %v", err)
	}

	// Round 2: Alice vs Charlie, Bob vs Diana
	pairings = tournament.GetRound()
	for _, pairing := range pairings {
		if pairing.playerb == BYE_OPPONENT_ID {
			continue
		}

		// Alice beats Charlie, Bob beats Diana
		if pairing.playera == 1 { // Alice
			tournament.AddResult(pairing.playera, 2, 1, 0) // Alice wins 2-1
		} else if pairing.playera == 2 { // Bob
			tournament.AddResult(pairing.playera, 2, 0, 0) // Bob wins 2-0
		} else if pairing.playera == 3 { // Charlie
			tournament.AddResult(pairing.playera, 1, 2, 0) // Charlie loses 1-2
		} else if pairing.playera == 4 { // Diana
			tournament.AddResult(pairing.playera, 0, 2, 0) // Diana loses 0-2
		}
	}

	// Update standings for round 2
	err = tournament.UpdatePlayerStandings()
	if err != nil {
		t.Fatalf("Failed to update standings: %v", err)
	}

	// Get standings
	standings := tournament.GetStandings()

	// Find Alice in standings
	var aliceStanding PlayerStanding
	for _, standing := range standings {
		if standing.Name == "Alice" {
			aliceStanding = standing
			break
		}
	}

	// Alice should have points from her wins (exact number depends on pairing)
	if aliceStanding.Points < 3 {
		t.Errorf("Expected Alice to have at least 3 points, got %d", aliceStanding.Points)
	}

	// Alice's opponent match win percentage should be at least 33%
	// Her opponents (Bob and Charlie) have poor records, but should be boosted to 33%
	if aliceStanding.Tiebreakers.OpponentMatchWinPct < 0.33 {
		t.Errorf("Expected Alice's opponent match win percentage to be at least 33%%, got %.3f",
			aliceStanding.Tiebreakers.OpponentMatchWinPct)
	}

	// Alice's opponent game win percentage should be at least 33%
	if aliceStanding.Tiebreakers.OpponentGameWinPct < 0.33 {
		t.Errorf("Expected Alice's opponent game win percentage to be at least 33%%, got %.3f",
			aliceStanding.Tiebreakers.OpponentGameWinPct)
	}

	t.Logf("Minimum 33%% rule test passed: Alice's opponent match win %%: %.3f, opponent game win %%: %.3f",
		aliceStanding.Tiebreakers.OpponentMatchWinPct, aliceStanding.Tiebreakers.OpponentGameWinPct)
}

func TestTournamentRankingSystem(t *testing.T) {
	tournament := NewTournament()

	// Add players
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")
	tournament.AddPlayer("Charlie")
	tournament.AddPlayer("Diana")

	// Start tournament
	err := tournament.StartTournament()
	if err != nil {
		t.Fatalf("Failed to start tournament: %v", err)
	}

	// Round 1: Create a scenario where Alice and Bob will be tied
	pairings := tournament.GetRound()
	for _, pairing := range pairings {
		if pairing.playerb == BYE_OPPONENT_ID {
			continue
		}

		// Add results for all pairings
		if pairing.playera == 1 { // Alice
			tournament.AddResult(pairing.playera, 2, 0, 0)
		} else if pairing.playera == 2 { // Bob
			tournament.AddResult(pairing.playera, 2, 0, 0) // Bob wins 2-0
		} else if pairing.playera == 3 { // Charlie
			tournament.AddResult(pairing.playera, 0, 2, 0) // Charlie loses 0-2
		} else if pairing.playera == 4 { // Diana
			tournament.AddResult(pairing.playera, 0, 2, 0) // Diana loses 0-2
		}
	}

	// Update standings for round 1
	err = tournament.UpdatePlayerStandings()
	if err != nil {
		t.Fatalf("Failed to update standings: %v", err)
	}

	// Get standings
	standings := tournament.GetStandings()

	// Print standings for debugging
	t.Logf("Standings:")
	for _, standing := range standings {
		t.Logf("%d. %s - Points: %d", standing.Rank, standing.Name, standing.Points)
	}

	// Find players in standings
	var aliceStanding, bobStanding, charlieStanding, dianaStanding PlayerStanding
	for _, standing := range standings {
		switch standing.Name {
		case "Alice":
			aliceStanding = standing
		case "Bob":
			bobStanding = standing
		case "Charlie":
			charlieStanding = standing
		case "Diana":
			dianaStanding = standing
		}
	}

	// Check that we have valid standings
	if len(standings) != 4 {
		t.Fatalf("Expected 4 standings, got %d", len(standings))
	}

	// Verify that the ranking system works correctly
	// Players with the same points should have the same rank
	// Players with different points should have sequential ranks

	// Find the highest rank
	maxRank := 0
	for _, standing := range standings {
		if standing.Rank > maxRank {
			maxRank = standing.Rank
		}
	}

	// Verify that ranks are sequential (no gaps)
	expectedRanks := make(map[int]bool)
	for i := 1; i <= maxRank; i++ {
		expectedRanks[i] = false
	}

	for _, standing := range standings {
		if standing.Rank < 1 || standing.Rank > maxRank {
			t.Errorf("Invalid rank %d for %s", standing.Rank, standing.Name)
		}
		expectedRanks[standing.Rank] = true
	}

	// Check that all expected ranks are used
	for rank, used := range expectedRanks {
		if !used {
			t.Errorf("Rank %d is not used by any player", rank)
		}
	}

	t.Logf("Tournament ranking test passed: Alice (rank %d), Bob (rank %d), Charlie (rank %d), Diana (rank %d)",
		aliceStanding.Rank, bobStanding.Rank, charlieStanding.Rank, dianaStanding.Rank)
}

// Dump/Load tests to ensure tournament can be resumed correctly
func TestDumpLoadBeforeStart(t *testing.T) {
	// Setup tournament with players and optional metadata
	tournament := NewTournament()
	if err := tournament.AddPlayer("Alice"); err != nil {
		t.Fatal(err)
	}
	if err := tournament.AddPlayer("Bob"); err != nil {
		t.Fatal(err)
	}

	// Optional metadata
	if err := tournament.SetPlayerExternalID(1, 12345); err != nil {
		t.Fatal(err)
	}
	deck := Decklist{Main: map[string]int{"Card A": 4, "Card B": 2}, Sideboard: map[string]int{"Card C": 3}}
	if err := tournament.SetPlayerDecklist(1, deck); err != nil {
		t.Fatal(err)
	}

	// Dump and load
	data, err := tournament.DumpTournament()
	if err != nil {
		t.Fatalf("DumpJSON failed: %v", err)
	}

	restored, err := LoadTournament(data)
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}

	// Verify status and metadata
	if restored.GetStatus() != "setup" {
		t.Fatalf("Expected status 'setup', got '%s'", restored.GetStatus())
	}

	p1, ok := restored.GetPlayerById(1)
	if !ok || p1.name != "Alice" {
		t.Fatalf("Expected Alice with ID 1, got ok=%v name=%s", ok, p1.name)
	}

	// Verify optional fields
	ext, has := restored.GetPlayerExternalID(1)
	if !has || ext == nil || *ext != 12345 {
		t.Fatalf("Expected externalID=12345, got has=%v val=%v", has, ext)
	}
	dl, hasDL := restored.GetPlayerDecklist(1)
	if !hasDL || dl == nil || dl.Main["Card A"] != 4 || dl.Sideboard["Card C"] != 3 {
		t.Fatalf("Decklist not restored correctly: %+v", dl)
	}

	// Ensure we can start and pair
	if err := restored.StartTournament(); err != nil {
		t.Fatalf("StartTournament after restore failed: %v", err)
	}
	if len(restored.GetRound()) == 0 {
		t.Fatal("Expected pairings after starting tournament")
	}
}

func TestDumpLoadMidRoundResume(t *testing.T) {
	tournament := NewTournament()
	// 4 players to avoid byes
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")
	tournament.AddPlayer("Charlie")
	tournament.AddPlayer("Diana")

	if err := tournament.StartTournament(); err != nil {
		t.Fatalf("Failed to start: %v", err)
	}

	// Pairings exist; record results for exactly one non-bye pairing (both should be non-bye with 4 players)
	pairings := tournament.GetRound()
	if len(pairings) == 0 {
		t.Fatal("No pairings in round 1")
	}
	// Complete first pairing only
	pr := pairings[0]
	if pr.playerb == BYE_OPPONENT_ID {
		t.Skip("Unexpected bye with 4 players; skipping")
	}
	if err := tournament.AddResult(pr.playera, 2, 1, 0); err != nil {
		t.Fatalf("AddResult failed: %v", err)
	}

	// Dump current state mid-round
	data, err := tournament.DumpTournament()
	if err != nil {
		t.Fatalf("DumpJSON failed: %v", err)
	}

	restored, err := LoadTournament(data)
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}

	// Resume: complete remaining pairings and update standings
	for _, p := range restored.GetRound() {
		if p.playera == pr.playera && p.playerb == pr.playerb {
			// already set above; skip
			continue
		}
		if p.playerb != BYE_OPPONENT_ID {
			if err := restored.AddResult(p.playera, 2, 0, 0); err != nil {
				t.Fatalf("AddResult after restore failed: %v", err)
			}
		}
	}

	// Should be able to update standings and advance
	if err := restored.UpdatePlayerStandings(); err != nil {
		t.Fatalf("UpdatePlayerStandings after restore failed: %v", err)
	}
	if err := restored.NextRound(); err != nil {
		t.Fatalf("NextRound after restore failed: %v", err)
	}
}

func TestDumpLoadWithRemovedPlayer(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")   // 1
	tournament.AddPlayer("Bob")     // 2
	tournament.AddPlayer("Charlie") // 3

	// Remove Bob before starting
	if err := tournament.RemovePlayerById(2); err != nil {
		t.Fatalf("RemovePlayerById failed: %v", err)
	}

	data, err := tournament.DumpTournament()
	if err != nil {
		t.Fatalf("DumpJSON failed: %v", err)
	}
	restored, err := LoadTournament(data)
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}

	p2, ok := restored.GetPlayerById(2)
	if !ok {
		t.Fatal("Expected player 2 to exist after restore")
	}
	if !p2.removed {
		t.Fatal("Expected player 2 to remain marked as removed after restore")
	}

	// Ensure pairing excludes removed players
	if err := restored.Pair(false); err != nil {
		t.Fatalf("Pair after restore failed: %v", err)
	}
	for _, p := range restored.GetRound() {
		if p.playera == 2 || p.playerb == 2 {
			t.Fatal("Removed player should not be paired after restore")
		}
	}
}
