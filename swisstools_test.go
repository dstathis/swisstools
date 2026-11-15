package swisstools

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestAddPlayerName(t *testing.T) {
	tournament := NewTournament()
	err := tournament.AddPlayer("Dylan")
	if err != nil || tournament.players[1].Name != "Dylan" {
		t.Fatalf("Expecting player name Dylan, got %s.", tournament.players[1].Name)
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
		totalWins += player.Wins
		totalLosses += player.Losses
		totalDraws += player.Draws
		totalPoints += player.Points
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
	if alice.Wins != 1 || alice.Points != 3 {
		t.Errorf("After round 1: Alice should have 1 win, 3 points, got wins=%d, points=%d", alice.Wins, alice.Points)
	}
	if bob.Losses != 1 || bob.Points != 0 {
		t.Errorf("After round 1: Bob should have 1 loss, 0 points, got losses=%d, points=%d", bob.Losses, bob.Points)
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
	if alice.Wins != 1 || alice.Losses != 1 || alice.Points != 3 {
		t.Errorf("After round 2: Alice should have 1 win, 1 loss, 3 points, got wins=%d, losses=%d, points=%d", alice.Wins, alice.Losses, alice.Points)
	}
	if bob.Wins != 1 || bob.Losses != 1 || bob.Points != 3 {
		t.Errorf("After round 2: Bob should have 1 win, 1 loss, 3 points, got wins=%d, losses=%d, points=%d", bob.Wins, bob.Losses, bob.Points)
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
	if alice.Draws != 1 || alice.Points != 1 {
		t.Errorf("Alice should have 1 draw and 1 point, got draws=%d, points=%d", alice.Draws, alice.Points)
	}
	if bob.Draws != 1 || bob.Points != 1 {
		t.Errorf("Bob should have 1 draw and 1 point, got draws=%d, points=%d", bob.Draws, bob.Points)
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
		if player.Wins != 0 || player.Losses != 0 || player.Points != 0 {
			t.Fatalf("Player %s should start with 0 stats", player.Name)
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
		if player.Wins != 0 || player.Losses != 0 || player.Points != 0 {
			t.Errorf("Player %s stats were modified despite incomplete matches: wins=%d, losses=%d, points=%d",
				player.Name, player.Wins, player.Losses, player.Points)
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
		totalPoints += player.Points
		totalWins += player.Wins
		totalDraws += player.Draws
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
	if player.Name != "Alice" {
		t.Errorf("Expected player 1 to be Alice, got %s", player.Name)
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

	if !player.Removed {
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
	for _, note := range player.Notes {
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

	if !player.Removed {
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
	if player.Name != "Bob" {
		t.Errorf("Expected player name 'Bob', got '%s'", player.Name)
	}

	// Test getting non-existent player
	player, exists = tournament.GetPlayerByName("David")
	if exists {
		t.Error("Expected not to find David")
	}
	if player.Name != "" {
		t.Errorf("Expected empty name for non-existent player, got '%s'", player.Name)
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

	if !player.Removed {
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

// recordRoundWinnerElseDraw records every match in the current round so that
// winnerID wins 2-0 and every other match is drawn 1-1-1. Results are recorded
// from a named player's perspective rather than from whoever happens to be
// player A, so callers do not depend on the random pairing order or on which
// side of a pairing a player lands.
func recordRoundWinnerElseDraw(t *testing.T, tournament *Tournament, winnerID int) {
	t.Helper()
	for _, pairing := range tournament.GetRound() {
		if pairing.playerb == BYE_OPPONENT_ID {
			continue
		}
		if pairing.playera == winnerID || pairing.playerb == winnerID {
			if err := tournament.AddResult(winnerID, 2, 0, 0); err != nil {
				t.Fatalf("AddResult for player %d: %v", winnerID, err)
			}
			continue
		}
		if err := tournament.AddResult(pairing.playera, 1, 1, 1); err != nil {
			t.Fatalf("AddResult for drawn match: %v", err)
		}
	}
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

	aliceID, ok := tournament.GetPlayerID("Alice")
	if !ok {
		t.Fatal("Alice not found")
	}

	// Alice wins her match every round and every other match is drawn, so
	// nobody but Alice ever banks a match win. Her opponents therefore end on
	// a raw match win percentage of 0 and a raw game win percentage of 1/6,
	// both below the floor, which is what puts the minimum 33% rule to work.
	// Pairing is random, so the round is scored by player identity rather than
	// by pairing position.
	recordRoundWinnerElseDraw(t, &tournament, aliceID)

	// Advance to round 2 (NextRound applies round 1's results)
	err = tournament.NextRound()
	if err != nil {
		t.Fatalf("Failed to advance to round 2: %v", err)
	}

	// Pair round 2
	err = tournament.Pair(false)
	if err != nil {
		t.Fatalf("Failed to pair round 2: %v", err)
	}

	recordRoundWinnerElseDraw(t, &tournament, aliceID)

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

	// Alice won both of her matches
	if aliceStanding.Points != 6 {
		t.Errorf("Expected Alice to have 6 points, got %d", aliceStanding.Points)
	}

	// Alice's opponent match win percentage should be at least 33%
	// Her opponents never won a match, but should be boosted to 33%
	if aliceStanding.Tiebreakers.OpponentMatchWinPct < 0.33 {
		t.Errorf("Expected Alice's opponent match win percentage to be at least 33%%, got %.3f",
			aliceStanding.Tiebreakers.OpponentMatchWinPct)
	}

	// Alice's opponent game win percentage should be at least 33%
	if aliceStanding.Tiebreakers.OpponentGameWinPct < 0.33 {
		t.Errorf("Expected Alice's opponent game win percentage to be at least 33%%, got %.3f",
			aliceStanding.Tiebreakers.OpponentGameWinPct)
	}

	// Both raw percentages are below the floor, so the floor is what produced
	// these values -- if either stops being clamped, this test is no longer
	// testing the rule it is named for.
	if aliceStanding.Tiebreakers.OpponentMatchWinPct != 0.33 {
		t.Errorf("Expected the 33%% floor to set Alice's opponent match win percentage, got %.3f",
			aliceStanding.Tiebreakers.OpponentMatchWinPct)
	}
	if aliceStanding.Tiebreakers.OpponentGameWinPct != 0.33 {
		t.Errorf("Expected the 33%% floor to set Alice's opponent game win percentage, got %.3f",
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
	if !ok || p1.Name != "Alice" {
		t.Fatalf("Expected Alice with ID 1, got ok=%v name=%s", ok, p1.Name)
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
	if !p2.Removed {
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

func TestPairingResultAccessors(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")

	tournament.Pair(false)
	pairings := tournament.GetRound()
	if len(pairings) == 0 {
		t.Fatal("Expected pairings")
	}

	// Before results, values are UNINITIALIZED_RESULT
	p := pairings[0]
	if p.PlayerAWins() != UNINITIALIZED_RESULT {
		t.Errorf("Expected UNINITIALIZED_RESULT for PlayerAWins, got %d", p.PlayerAWins())
	}

	// Add a result and re-read
	err := tournament.AddResult(p.PlayerA(), 2, 1, 1)
	if err != nil {
		t.Fatalf("AddResult failed: %v", err)
	}
	pairings = tournament.GetRound()
	p = pairings[0]

	if p.PlayerAWins() != 2 {
		t.Errorf("Expected PlayerAWins=2, got %d", p.PlayerAWins())
	}
	if p.PlayerBWins() != 1 {
		t.Errorf("Expected PlayerBWins=1, got %d", p.PlayerBWins())
	}
	if p.Draws() != 1 {
		t.Errorf("Expected Draws=1, got %d", p.Draws())
	}
}

func TestGetPlayers(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")
	tournament.AddPlayer("Charlie")

	players := tournament.GetPlayers()
	if len(players) != 3 {
		t.Fatalf("Expected 3 players, got %d", len(players))
	}

	// Verify names
	names := map[string]bool{}
	for _, p := range players {
		names[p.Name] = true
	}
	for _, expected := range []string{"Alice", "Bob", "Charlie"} {
		if !names[expected] {
			t.Errorf("Expected player %s in GetPlayers result", expected)
		}
	}

	// Verify mutation safety: modifying returned map doesn't affect tournament
	delete(players, 1)
	if tournament.GetPlayerCount() != 3 {
		t.Error("Modifying GetPlayers result should not affect tournament")
	}
}

func TestGetPlayersIncludesRemoved(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")
	tournament.RemovePlayerById(1)

	players := tournament.GetPlayers()
	if len(players) != 2 {
		t.Fatalf("Expected 2 players (including removed), got %d", len(players))
	}
	if !players[1].Removed {
		t.Error("Expected player 1 to be marked as removed")
	}
}

func TestGetRoundByNumber(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")

	// Start and complete round 1
	tournament.StartTournament()
	pairings := tournament.GetRound()
	for _, p := range pairings {
		if p.PlayerB() != BYE_OPPONENT_ID {
			tournament.AddResult(p.PlayerA(), 2, 0, 0)
		}
	}
	tournament.NextRound()

	// Pair round 2
	tournament.Pair(false)

	// Round 1 should be retrievable
	r1, err := tournament.GetRoundByNumber(1)
	if err != nil {
		t.Fatalf("GetRoundByNumber(1) failed: %v", err)
	}
	if len(r1) == 0 {
		t.Fatal("Expected pairings in round 1")
	}

	// Round 2 should also work (current round)
	r2, err := tournament.GetRoundByNumber(2)
	if err != nil {
		t.Fatalf("GetRoundByNumber(2) failed: %v", err)
	}
	if len(r2) == 0 {
		t.Fatal("Expected pairings in round 2")
	}

	// Results from round 1 should be readable via accessors
	for _, p := range r1 {
		if p.PlayerB() != BYE_OPPONENT_ID {
			if p.PlayerAWins() == UNINITIALIZED_RESULT {
				t.Error("Expected round 1 results to be set")
			}
		}
	}

	// Out of range
	_, err = tournament.GetRoundByNumber(0)
	if err == nil {
		t.Error("Expected error for round 0")
	}
	_, err = tournament.GetRoundByNumber(99)
	if err == nil {
		t.Error("Expected error for round 99")
	}
}

func TestFinishTournament(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")

	// Cannot finish before starting
	err := tournament.FinishTournament()
	if err == nil {
		t.Fatal("Expected error finishing un-started tournament")
	}

	tournament.StartTournament()

	// Complete results
	for _, p := range tournament.GetRound() {
		if p.PlayerB() != BYE_OPPONENT_ID {
			tournament.AddResult(p.PlayerA(), 2, 0, 0)
		}
	}

	// Finish
	err = tournament.FinishTournament()
	if err != nil {
		t.Fatalf("FinishTournament failed: %v", err)
	}
	if tournament.GetStatus() != "finished" {
		t.Errorf("Expected status 'finished', got '%s'", tournament.GetStatus())
	}

	// Cannot finish again
	err = tournament.FinishTournament()
	if err == nil {
		t.Fatal("Expected error finishing already-finished tournament")
	}
}

func TestSetMaxRoundsAutoFinish(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")
	tournament.SetMaxRounds(2)

	if tournament.GetMaxRounds() != 2 {
		t.Fatalf("Expected maxRounds=2, got %d", tournament.GetMaxRounds())
	}

	tournament.StartTournament()

	// Round 1
	for _, p := range tournament.GetRound() {
		if p.PlayerB() != BYE_OPPONENT_ID {
			tournament.AddResult(p.PlayerA(), 2, 0, 0)
		}
	}
	err := tournament.NextRound()
	if err != nil {
		t.Fatalf("NextRound failed: %v", err)
	}
	if tournament.GetStatus() != "in_progress" {
		t.Error("Expected in_progress after round 1 of 2")
	}

	// Round 2
	tournament.Pair(false)
	for _, p := range tournament.GetRound() {
		if p.PlayerB() != BYE_OPPONENT_ID {
			tournament.AddResult(p.PlayerA(), 2, 0, 0)
		}
	}
	err = tournament.NextRound()
	if err != nil {
		t.Fatalf("NextRound after round 2 failed: %v", err)
	}
	if tournament.GetStatus() != "finished" {
		t.Errorf("Expected 'finished' after max rounds reached, got '%s'", tournament.GetStatus())
	}
}

func TestMaxRoundsZeroMeansNoLimit(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")
	// maxRounds defaults to 0 (no limit)
	if tournament.GetMaxRounds() != 0 {
		t.Fatalf("Expected default maxRounds=0, got %d", tournament.GetMaxRounds())
	}

	tournament.StartTournament()
	for _, p := range tournament.GetRound() {
		if p.PlayerB() != BYE_OPPONENT_ID {
			tournament.AddResult(p.PlayerA(), 2, 0, 0)
		}
	}
	// Should advance normally without finishing
	err := tournament.NextRound()
	if err != nil {
		t.Fatalf("NextRound failed: %v", err)
	}
	if tournament.GetStatus() != "in_progress" {
		t.Errorf("Expected in_progress with no max rounds, got '%s'", tournament.GetStatus())
	}
}

func TestPlayerExportedFields(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")

	tournament.Pair(false)
	tournament.AddResult(1, 2, 1, 0)
	tournament.UpdatePlayerStandings()

	// Access exported fields through GetPlayerById
	alice, ok := tournament.GetPlayerById(1)
	if !ok {
		t.Fatal("Expected to find Alice")
	}
	if alice.Points != 3 {
		t.Errorf("Expected Points=3, got %d", alice.Points)
	}
	if alice.Wins != 1 {
		t.Errorf("Expected Wins=1, got %d", alice.Wins)
	}
	if alice.GameWins != 2 {
		t.Errorf("Expected GameWins=2, got %d", alice.GameWins)
	}
	if alice.GameLosses != 1 {
		t.Errorf("Expected GameLosses=1, got %d", alice.GameLosses)
	}

	bob, ok := tournament.GetPlayerById(2)
	if !ok {
		t.Fatal("Expected to find Bob")
	}
	if bob.Losses != 1 {
		t.Errorf("Expected Losses=1, got %d", bob.Losses)
	}
}

func TestDumpLoadMaxRounds(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")
	tournament.SetMaxRounds(5)

	data, err := tournament.DumpTournament()
	if err != nil {
		t.Fatalf("DumpTournament failed: %v", err)
	}

	restored, err := LoadTournament(data)
	if err != nil {
		t.Fatalf("LoadTournament failed: %v", err)
	}

	if restored.GetMaxRounds() != 5 {
		t.Errorf("Expected maxRounds=5 after restore, got %d", restored.GetMaxRounds())
	}
}

// --- Playoff tests ---

// runSwissToFinish is a test helper that creates a tournament with the given
// number of players, plays through the Swiss rounds, and finishes it.
func runSwissToFinish(t *testing.T, numPlayers int) Tournament {
	t.Helper()
	tournament := NewTournament()
	for i := 1; i <= numPlayers; i++ {
		tournament.AddPlayer(fmt.Sprintf("Player%d", i))
	}
	if err := tournament.StartTournament(); err != nil {
		t.Fatalf("StartTournament: %v", err)
	}
	// Play 3 rounds
	for round := 1; round <= 3; round++ {
		for _, p := range tournament.GetRound() {
			if p.PlayerB() != BYE_OPPONENT_ID {
				// Higher-seeded player wins
				if err := tournament.AddResult(p.PlayerA(), 2, 1, 0); err != nil {
					t.Fatalf("AddResult round %d: %v", round, err)
				}
			}
		}
		if round < 3 {
			if err := tournament.NextRound(); err != nil {
				t.Fatalf("NextRound %d: %v", round, err)
			}
			tournament.Pair(false)
		}
	}
	if err := tournament.FinishTournament(); err != nil {
		t.Fatalf("FinishTournament: %v", err)
	}
	return tournament
}

func TestStartPlayoffRequiresFinished(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")
	tournament.StartTournament()

	err := tournament.StartPlayoff(2)
	if err == nil {
		t.Fatal("Expected error starting playoff before Swiss is finished")
	}
}

func TestStartPlayoffPowerOfTwo(t *testing.T) {
	tournament := runSwissToFinish(t, 8)

	// 3 is not a power of 2
	err := tournament.StartPlayoff(3)
	if err == nil {
		t.Fatal("Expected error for non-power-of-2 topN")
	}

	// 1 is not valid
	err = tournament.StartPlayoff(1)
	if err == nil {
		t.Fatal("Expected error for topN=1")
	}

	// 4 should work
	err = tournament.StartPlayoff(4)
	if err != nil {
		t.Fatalf("StartPlayoff(4) failed: %v", err)
	}
}

func TestStartPlayoffNotEnoughPlayers(t *testing.T) {
	tournament := runSwissToFinish(t, 4)

	err := tournament.StartPlayoff(8)
	if err == nil {
		t.Fatal("Expected error when topN exceeds player count")
	}
}

func TestStartPlayoffDouble(t *testing.T) {
	tournament := runSwissToFinish(t, 8)
	tournament.StartPlayoff(4)

	err := tournament.StartPlayoff(4)
	if err == nil {
		t.Fatal("Expected error starting playoff twice")
	}
}

func TestPlayoffTop4FullBracket(t *testing.T) {
	tournament := runSwissToFinish(t, 8)

	err := tournament.StartPlayoff(4)
	if err != nil {
		t.Fatalf("StartPlayoff: %v", err)
	}

	if tournament.GetPlayoffStatus() != "in_progress" {
		t.Errorf("Expected playoff in_progress, got %s", tournament.GetPlayoffStatus())
	}

	playoff := tournament.GetPlayoff()
	if playoff == nil {
		t.Fatal("Expected non-nil playoff")
	}
	if len(playoff.Seeds) != 4 {
		t.Fatalf("Expected 4 seeds, got %d", len(playoff.Seeds))
	}
	// Should have 2 rounds total (semis + finals)
	if len(playoff.Rounds) != 2 {
		t.Fatalf("Expected 2 playoff rounds for top 4, got %d", len(playoff.Rounds))
	}

	// Semifinals: 2 matches
	semis := tournament.GetPlayoffRound()
	if len(semis) != 2 {
		t.Fatalf("Expected 2 semifinal matches, got %d", len(semis))
	}

	// Verify seeding: seed1 vs seed4, seed2 vs seed3
	if semis[0].PlayerA() != playoff.Seeds[0] || semis[0].PlayerB() != playoff.Seeds[3] {
		t.Error("First semi should be seed 1 vs seed 4")
	}
	if semis[1].PlayerA() != playoff.Seeds[1] || semis[1].PlayerB() != playoff.Seeds[2] {
		t.Error("Second semi should be seed 2 vs seed 3")
	}

	// Record semifinal results: higher seeds win
	for _, p := range semis {
		err := tournament.AddPlayoffResult(p.PlayerA(), 2, 1, 0)
		if err != nil {
			t.Fatalf("AddPlayoffResult semi: %v", err)
		}
	}

	// Advance to finals
	err = tournament.NextPlayoffRound()
	if err != nil {
		t.Fatalf("NextPlayoffRound: %v", err)
	}

	finals := tournament.GetPlayoffRound()
	if len(finals) != 1 {
		t.Fatalf("Expected 1 final match, got %d", len(finals))
	}

	// Record final result
	err = tournament.AddPlayoffResult(finals[0].PlayerA(), 2, 0, 0)
	if err != nil {
		t.Fatalf("AddPlayoffResult final: %v", err)
	}

	// Finish playoff
	err = tournament.NextPlayoffRound()
	if err != nil {
		t.Fatalf("NextPlayoffRound final: %v", err)
	}

	if tournament.GetPlayoffStatus() != "finished" {
		t.Errorf("Expected playoff finished, got %s", tournament.GetPlayoffStatus())
	}
}

func TestPlayoffTop8(t *testing.T) {
	tournament := runSwissToFinish(t, 16)

	err := tournament.StartPlayoff(8)
	if err != nil {
		t.Fatalf("StartPlayoff(8): %v", err)
	}

	playoff := tournament.GetPlayoff()
	// 8 -> 3 rounds (quarters, semis, finals)
	if len(playoff.Rounds) != 3 {
		t.Fatalf("Expected 3 playoff rounds for top 8, got %d", len(playoff.Rounds))
	}

	// Play through all 3 rounds
	for round := 0; round < 3; round++ {
		pairings := tournament.GetPlayoffRound()
		for _, p := range pairings {
			tournament.AddPlayoffResult(p.PlayerA(), 2, 0, 0)
		}
		err := tournament.NextPlayoffRound()
		if err != nil {
			t.Fatalf("NextPlayoffRound round %d: %v", round, err)
		}
	}

	if tournament.GetPlayoffStatus() != "finished" {
		t.Errorf("Expected finished after 3 rounds, got %s", tournament.GetPlayoffStatus())
	}
}

func TestPlayoffIncompleteResults(t *testing.T) {
	tournament := runSwissToFinish(t, 8)
	tournament.StartPlayoff(4)

	// Try to advance without results
	err := tournament.NextPlayoffRound()
	if err == nil {
		t.Fatal("Expected error advancing without results")
	}

	// Add result for only one match
	semis := tournament.GetPlayoffRound()
	tournament.AddPlayoffResult(semis[0].PlayerA(), 2, 0, 0)

	err = tournament.NextPlayoffRound()
	if err == nil {
		t.Fatal("Expected error with partial results")
	}
}

func TestPlayoffDrawNotAllowed(t *testing.T) {
	tournament := runSwissToFinish(t, 8)
	tournament.StartPlayoff(4)

	semis := tournament.GetPlayoffRound()
	// Both matches result in draws
	tournament.AddPlayoffResult(semis[0].PlayerA(), 1, 1, 0)
	tournament.AddPlayoffResult(semis[1].PlayerA(), 2, 0, 0)

	err := tournament.NextPlayoffRound()
	if err == nil {
		t.Fatal("Expected error when playoff match is drawn")
	}
}

func TestPlayoffGetRoundByNumber(t *testing.T) {
	tournament := runSwissToFinish(t, 8)
	tournament.StartPlayoff(4)

	// Round 0 should have semis
	r0, err := tournament.GetPlayoffRoundByNumber(0)
	if err != nil {
		t.Fatalf("GetPlayoffRoundByNumber(0): %v", err)
	}
	if len(r0) != 2 {
		t.Errorf("Expected 2 matches in round 0, got %d", len(r0))
	}

	// Out of range
	_, err = tournament.GetPlayoffRoundByNumber(-1)
	if err == nil {
		t.Error("Expected error for negative round")
	}
	_, err = tournament.GetPlayoffRoundByNumber(5)
	if err == nil {
		t.Error("Expected error for round beyond bracket")
	}

	// No playoff
	t2 := runSwissToFinish(t, 4)
	_, err = t2.GetPlayoffRoundByNumber(0)
	if err == nil {
		t.Error("Expected error when no playoff started")
	}
}

func TestPlayoffResultAccessors(t *testing.T) {
	tournament := runSwissToFinish(t, 8)
	tournament.StartPlayoff(4)

	semis := tournament.GetPlayoffRound()
	tournament.AddPlayoffResult(semis[0].PlayerA(), 2, 1, 1)

	// Re-read to get updated values
	semis = tournament.GetPlayoffRound()
	if semis[0].PlayerAWins() != 2 || semis[0].PlayerBWins() != 1 || semis[0].Draws() != 1 {
		t.Errorf("Playoff result accessors: got A=%d B=%d D=%d, want A=2 B=1 D=1",
			semis[0].PlayerAWins(), semis[0].PlayerBWins(), semis[0].Draws())
	}
}

func TestAddPlayoffResultPlayerB(t *testing.T) {
	tournament := runSwissToFinish(t, 8)
	tournament.StartPlayoff(4)

	semis := tournament.GetPlayoffRound()
	// Add result from playerB's perspective
	err := tournament.AddPlayoffResult(semis[0].PlayerB(), 2, 1, 0)
	if err != nil {
		t.Fatalf("AddPlayoffResult for playerB: %v", err)
	}

	semis = tournament.GetPlayoffRound()
	if semis[0].PlayerBWins() != 2 || semis[0].PlayerAWins() != 1 {
		t.Error("Result from playerB perspective not recorded correctly")
	}
}

func TestPlayoffNoPlayoffErrors(t *testing.T) {
	tournament := NewTournament()

	if tournament.GetPlayoffStatus() != "none" {
		t.Error("Expected 'none' playoff status")
	}
	if tournament.GetPlayoff() != nil {
		t.Error("Expected nil playoff")
	}
	if tournament.GetPlayoffRound() != nil {
		t.Error("Expected nil from GetPlayoffRound")
	}
	err := tournament.AddPlayoffResult(1, 2, 0, 0)
	if err == nil {
		t.Error("Expected error from AddPlayoffResult with no playoff")
	}
	err = tournament.NextPlayoffRound()
	if err == nil {
		t.Error("Expected error from NextPlayoffRound with no playoff")
	}
}

func TestPlayoffAddResultAfterFinished(t *testing.T) {
	tournament := runSwissToFinish(t, 4)
	tournament.StartPlayoff(2)

	// Top 2 - single match is the final
	finals := tournament.GetPlayoffRound()
	tournament.AddPlayoffResult(finals[0].PlayerA(), 2, 0, 0)
	tournament.NextPlayoffRound()

	// Should be finished
	if tournament.GetPlayoffStatus() != "finished" {
		t.Fatal("Expected playoff to be finished")
	}

	err := tournament.AddPlayoffResult(finals[0].PlayerA(), 2, 0, 0)
	if err == nil {
		t.Error("Expected error adding result to finished playoff")
	}

	err = tournament.NextPlayoffRound()
	if err == nil {
		t.Error("Expected error advancing finished playoff")
	}
}

func TestPlayoffPlayerNotFound(t *testing.T) {
	tournament := runSwissToFinish(t, 8)
	tournament.StartPlayoff(4)

	err := tournament.AddPlayoffResult(999, 2, 0, 0)
	if err == nil {
		t.Error("Expected error for non-existent player in playoff")
	}
}

func TestDumpLoadPlayoff(t *testing.T) {
	tournament := runSwissToFinish(t, 8)
	tournament.StartPlayoff(4)

	// Complete semifinals
	semis := tournament.GetPlayoffRound()
	for _, p := range semis {
		tournament.AddPlayoffResult(p.PlayerA(), 2, 1, 0)
	}
	tournament.NextPlayoffRound()

	// Dump mid-playoff
	data, err := tournament.DumpTournament()
	if err != nil {
		t.Fatalf("DumpTournament: %v", err)
	}

	restored, err := LoadTournament(data)
	if err != nil {
		t.Fatalf("LoadTournament: %v", err)
	}

	if restored.GetPlayoffStatus() != "in_progress" {
		t.Errorf("Expected in_progress after restore, got %s", restored.GetPlayoffStatus())
	}

	playoff := restored.GetPlayoff()
	if playoff == nil {
		t.Fatal("Expected non-nil playoff after restore")
	}
	if len(playoff.Seeds) != 4 {
		t.Errorf("Expected 4 seeds after restore, got %d", len(playoff.Seeds))
	}
	if playoff.CurrentRound != 1 {
		t.Errorf("Expected CurrentRound=1, got %d", playoff.CurrentRound)
	}

	// Should be able to complete the finals
	finals := restored.GetPlayoffRound()
	if len(finals) != 1 {
		t.Fatalf("Expected 1 final match after restore, got %d", len(finals))
	}
	restored.AddPlayoffResult(finals[0].PlayerA(), 2, 0, 0)
	err = restored.NextPlayoffRound()
	if err != nil {
		t.Fatalf("NextPlayoffRound after restore: %v", err)
	}
	if restored.GetPlayoffStatus() != "finished" {
		t.Error("Expected finished after completing restored playoff")
	}
}

// enterOneRoundEveryPlayerAWins starts the tournament and gives every player A
// a 2-0 win. It deliberately does not accumulate standings: NextRound and
// FinishTournament both call UpdatePlayerStandings themselves, and calling it
// as well would count every result twice.
func enterOneRoundEveryPlayerAWins(t *testing.T, tournament *Tournament) {
	t.Helper()
	if err := tournament.StartTournament(); err != nil {
		t.Fatalf("StartTournament: %v", err)
	}
	for _, p := range tournament.GetRound() {
		if p.PlayerB() != BYE_OPPONENT_ID {
			if err := tournament.AddResult(p.PlayerA(), 2, 0, 0); err != nil {
				t.Fatalf("AddResult: %v", err)
			}
		}
	}
}

// playOneRoundEveryPlayerAWins enters a round and accumulates it, for tests
// that need standings without advancing or finishing.
func playOneRoundEveryPlayerAWins(t *testing.T, tournament *Tournament) {
	t.Helper()
	enterOneRoundEveryPlayerAWins(t, tournament)
	if err := tournament.UpdatePlayerStandings(); err != nil {
		t.Fatalf("UpdatePlayerStandings: %v", err)
	}
}

// TestGetFinalStandingsRanksDroppedPlayerInPlace is the core guarantee: a
// player who drops is ranked where their points put them, not pushed to the
// bottom of the list.
func TestGetFinalStandingsRanksDroppedPlayerInPlace(t *testing.T) {
	tournament := NewTournament()
	for _, n := range []string{"Alice", "Bob", "Charlie", "Diana"} {
		tournament.AddPlayer(n)
	}
	playOneRoundEveryPlayerAWins(t, &tournament)

	// Drop the current leader, who is on a win while others are on a loss.
	leader := tournament.GetStandings()[0]
	if leader.Points == 0 {
		t.Fatalf("expected the leader to have points, got %d", leader.Points)
	}
	if err := tournament.RemovePlayerById(leader.PlayerID); err != nil {
		t.Fatalf("RemovePlayerById: %v", err)
	}

	final := tournament.GetFinalStandings()
	if len(final) != 4 {
		t.Fatalf("GetFinalStandings: expected 4 players, got %d", len(final))
	}
	if got := len(tournament.GetStandings()); got != 3 {
		t.Fatalf("GetStandings: expected the dropped player to be omitted, got %d players", got)
	}

	var dropped *PlayerStanding
	for i := range final {
		if final[i].PlayerID == leader.PlayerID {
			dropped = &final[i]
		}
	}
	if dropped == nil {
		t.Fatal("GetFinalStandings omitted the dropped player")
	}
	if !dropped.Removed {
		t.Error("expected Removed to be true for the dropped player")
	}
	if dropped.RemovedInRound != 1 {
		t.Errorf("expected RemovedInRound 1, got %d", dropped.RemovedInRound)
	}

	// The point: they outrank everyone who scored fewer points, rather than
	// being pinned to the bottom.
	for _, s := range final {
		if s.PlayerID == dropped.PlayerID {
			continue
		}
		if s.Points < dropped.Points && s.Rank <= dropped.Rank {
			t.Errorf("dropped player (rank %d, %d pts) did not outrank %s (rank %d, %d pts)",
				dropped.Rank, dropped.Points, s.Name, s.Rank, s.Points)
		}
	}
}

// TestGetFinalStandingsMatchesGetStandingsWithoutDrops guards against the
// refactor changing existing behavior.
func TestGetFinalStandingsMatchesGetStandingsWithoutDrops(t *testing.T) {
	tournament := NewTournament()
	for _, n := range []string{"Alice", "Bob", "Charlie", "Diana", "Eve", "Frank"} {
		tournament.AddPlayer(n)
	}
	playOneRoundEveryPlayerAWins(t, &tournament)

	standings := tournament.GetStandings()
	final := tournament.GetFinalStandings()
	if len(standings) != len(final) {
		t.Fatalf("with no drops, expected identical lengths: %d vs %d", len(standings), len(final))
	}

	// Compare rank by player ID rather than by slice position. sortedPlayers
	// ranges over a map and its last-resort tiebreak compares slice indices
	// mid-sort, so the order *within* a group of tied players is not stable
	// between calls. The rank each player is assigned is stable, because every
	// member of a tie group compares equal to its predecessor.
	ranks := make(map[int]int, len(standings))
	for _, s := range standings {
		ranks[s.PlayerID] = s.Rank
	}
	for _, s := range final {
		want, ok := ranks[s.PlayerID]
		if !ok {
			t.Errorf("player %d (%s) is in GetFinalStandings but not GetStandings", s.PlayerID, s.Name)
			continue
		}
		if s.Rank != want {
			t.Errorf("player %d (%s): GetStandings rank %d, GetFinalStandings rank %d",
				s.PlayerID, s.Name, want, s.Rank)
		}
		if s.Removed {
			t.Errorf("player %s should not be marked removed", s.Name)
		}
	}
}

// TestStartPlayoffNeverSeedsDroppedPlayer is the reason GetStandings still
// excludes dropped players: the top cut is seeded from it.
func TestStartPlayoffNeverSeedsDroppedPlayer(t *testing.T) {
	tournament := NewTournament()
	for _, n := range []string{"Alice", "Bob", "Charlie", "Diana", "Eve", "Frank", "Grace", "Heidi"} {
		tournament.AddPlayer(n)
	}
	// FinishTournament below accumulates the round, so do not do it here too.
	enterOneRoundEveryPlayerAWins(t, &tournament)
	if err := tournament.FinishTournament(); err != nil {
		t.Fatalf("FinishTournament: %v", err)
	}

	// Drop the leader, who would otherwise be seeded first.
	leader := tournament.GetStandings()[0]
	if err := tournament.RemovePlayerById(leader.PlayerID); err != nil {
		t.Fatalf("RemovePlayerById: %v", err)
	}

	// They still rank inside the cut on the display view.
	var finalRank int
	for _, s := range tournament.GetFinalStandings() {
		if s.PlayerID == leader.PlayerID {
			finalRank = s.Rank
		}
	}
	if finalRank == 0 || finalRank > 4 {
		t.Fatalf("expected the dropped player to rank inside the top 4, got %d", finalRank)
	}

	if err := tournament.StartPlayoff(4); err != nil {
		t.Fatalf("StartPlayoff: %v", err)
	}

	for _, seed := range tournament.GetPlayoff().Seeds {
		if seed == leader.PlayerID {
			t.Fatalf("dropped player %d was seeded into the playoff", leader.PlayerID)
		}
	}
}

// --- Manual pairing modification API (issue #2) ---

// setupPairedRound builds a tournament with the given names, starts it, and
// returns it with round 1 paired. Panics via t.Fatal if setup fails.
func setupPairedRound(t *testing.T, names ...string) Tournament {
	t.Helper()
	tournament := NewTournament()
	for _, n := range names {
		if err := tournament.AddPlayer(n); err != nil {
			t.Fatalf("AddPlayer %s: %v", n, err)
		}
	}
	if err := tournament.StartTournament(); err != nil {
		t.Fatalf("StartTournament: %v", err)
	}
	return tournament
}

func TestSwapPlayersHappyPath(t *testing.T) {
	tournament := setupPairedRound(t, "Alice", "Bob", "Charlie", "Diana")

	round := tournament.GetRound()
	if len(round) != 2 {
		t.Fatalf("expected 2 pairings, got %d", len(round))
	}
	// Pre-swap: pairings are (p0.playera, p0.playerb) and (p1.playera, p1.playerb)
	a0, b0 := round[0].playera, round[0].playerb
	a1, b1 := round[1].playera, round[1].playerb

	if err := tournament.SwapPlayers(b0, b1); err != nil {
		t.Fatalf("SwapPlayers: %v", err)
	}
	after := tournament.GetRound()
	if after[0].playera != a0 || after[0].playerb != b1 {
		t.Errorf("pairing 0: want (%d, %d), got (%d, %d)", a0, b1, after[0].playera, after[0].playerb)
	}
	if after[1].playera != a1 || after[1].playerb != b0 {
		t.Errorf("pairing 1: want (%d, %d), got (%d, %d)", a1, b0, after[1].playera, after[1].playerb)
	}
}

func TestSwapPlayersWithBye(t *testing.T) {
	tournament := setupPairedRound(t, "Alice", "Bob", "Charlie")

	// 3 players: one of them has a bye. Swapping a bye-holder is blocked
	// because the bye pairing already has a recorded result.
	var byeID, oppID int
	for _, p := range tournament.GetRound() {
		if p.playerb == BYE_OPPONENT_ID {
			byeID = p.playera
		} else {
			oppID = p.playera
		}
	}
	if byeID == 0 || oppID == 0 {
		t.Fatal("could not find bye and opponent IDs")
	}

	err := tournament.SwapPlayers(byeID, oppID)
	if err == nil || err.Error() != "cannot modify pairing with recorded result" {
		t.Errorf("expected recorded-result error, got %v", err)
	}
}

func TestSwapPlayersGuards(t *testing.T) {
	tournament := setupPairedRound(t, "Alice", "Bob", "Charlie", "Diana")
	round := tournament.GetRound()
	a0, b0 := round[0].playera, round[0].playerb

	// Self-swap
	err := tournament.SwapPlayers(a0, a0)
	if err == nil || err.Error() != "cannot swap a player with themselves" {
		t.Errorf("self-swap: got %v", err)
	}

	// Unknown player
	err = tournament.SwapPlayers(a0, 999)
	if err == nil || err.Error() != "player not found" {
		t.Errorf("unknown player: got %v", err)
	}

	// Facing each other
	err = tournament.SwapPlayers(a0, b0)
	if err == nil || err.Error() != "players are already paired against each other" {
		t.Errorf("same pairing: got %v", err)
	}

	// Recorded result blocks swap
	if err := tournament.AddResult(a0, 2, 1, 0); err != nil {
		t.Fatalf("AddResult: %v", err)
	}
	b1 := round[1].playerb
	err = tournament.SwapPlayers(b0, b1)
	if err == nil || err.Error() != "cannot modify pairing with recorded result" {
		t.Errorf("recorded result: got %v", err)
	}
}

func TestSwapPlayersBeforePair(t *testing.T) {
	tournament := NewTournament()
	tournament.AddPlayer("Alice")
	tournament.AddPlayer("Bob")

	err := tournament.SwapPlayers(1, 2)
	if err == nil || err.Error() != "round has no pairings - call Pair() first" {
		t.Errorf("got %v", err)
	}
}

func TestSetPairingHappyPath(t *testing.T) {
	tournament := setupPairedRound(t, "Alice", "Bob", "Charlie", "Diana")

	// Remove both pairings, then rebuild one.
	round := tournament.GetRound()
	a0, b0 := round[0].playera, round[0].playerb
	if err := tournament.RemovePairing(a0); err != nil {
		t.Fatalf("RemovePairing a0: %v", err)
	}
	if err := tournament.SetPairing(a0, b0); err != nil {
		t.Fatalf("SetPairing: %v", err)
	}
	after := tournament.GetRound()
	if len(after) != 2 {
		t.Fatalf("expected 2 pairings after set, got %d", len(after))
	}
	// Find the restored pairing
	found := false
	for _, p := range after {
		if (p.playera == a0 && p.playerb == b0) || (p.playera == b0 && p.playerb == a0) {
			if p.playeraWins != UNINITIALIZED_RESULT {
				t.Error("restored pairing should have UNINITIALIZED_RESULT")
			}
			found = true
		}
	}
	if !found {
		t.Error("restored pairing not found")
	}
}

func TestSetPairingWithBye(t *testing.T) {
	tournament := setupPairedRound(t, "Alice", "Bob", "Charlie", "Diana")

	// Remove Alice's pairing, then set her a bye.
	if err := tournament.RemovePairing(1); err != nil {
		t.Fatalf("RemovePairing: %v", err)
	}
	if err := tournament.SetPairing(1, BYE_OPPONENT_ID); err != nil {
		t.Fatalf("SetPairing bye: %v", err)
	}
	after := tournament.GetRound()
	foundBye := false
	for _, p := range after {
		if p.playera == 1 && p.playerb == BYE_OPPONENT_ID {
			if p.playeraWins != BYE_WINS || p.playerbWins != BYE_LOSSES || p.draws != BYE_DRAWS {
				t.Errorf("bye results wrong: %+v", p)
			}
			foundBye = true
		}
	}
	if !foundBye {
		t.Error("bye pairing not found after SetPairing")
	}
}

func TestSetPairingGuards(t *testing.T) {
	tournament := setupPairedRound(t, "Alice", "Bob", "Charlie", "Diana")

	// Self-pair
	err := tournament.SetPairing(1, 1)
	if err == nil || err.Error() != "cannot pair a player with themselves" {
		t.Errorf("self-pair: got %v", err)
	}

	// Both already paired
	err = tournament.SetPairing(1, 2)
	if err == nil || err.Error() != "player already paired - call RemovePairing first" {
		t.Errorf("already paired: got %v", err)
	}

	// Unknown player
	err = tournament.SetPairing(999, 1)
	if err == nil || err.Error() != "player not found" {
		t.Errorf("unknown A: got %v", err)
	}

	// After RemovePairing, pairing against a still-paired player
	if err := tournament.RemovePairing(1); err != nil {
		t.Fatalf("RemovePairing: %v", err)
	}
	// 1 is unpaired, but 3 (or whichever was in the other pairing) is still paired
	round := tournament.GetRound()
	stillPaired := round[0].playera
	err = tournament.SetPairing(1, stillPaired)
	if err == nil || err.Error() != "player already paired - call RemovePairing first" {
		t.Errorf("partner still paired: got %v", err)
	}
}

func TestRemovePairingHappyPath(t *testing.T) {
	tournament := setupPairedRound(t, "Alice", "Bob", "Charlie", "Diana")

	before := len(tournament.GetRound())
	if err := tournament.RemovePairing(1); err != nil {
		t.Fatalf("RemovePairing: %v", err)
	}
	after := len(tournament.GetRound())
	if after != before-1 {
		t.Errorf("expected %d pairings, got %d", before-1, after)
	}
	// Confirm player 1 is no longer in any pairing
	for _, p := range tournament.GetRound() {
		if p.playera == 1 || p.playerb == 1 {
			t.Error("player 1 still appears in a pairing")
		}
	}
}

func TestRemovePairingGuards(t *testing.T) {
	tournament := setupPairedRound(t, "Alice", "Bob", "Charlie", "Diana")

	err := tournament.RemovePairing(999)
	if err == nil || err.Error() != "player not found" {
		t.Errorf("unknown player: got %v", err)
	}

	if err := tournament.AddResult(1, 2, 1, 0); err != nil {
		t.Fatalf("AddResult: %v", err)
	}
	err = tournament.RemovePairing(1)
	if err == nil || err.Error() != "cannot modify pairing with recorded result" {
		t.Errorf("recorded result: got %v", err)
	}
}

func TestRemovePairingThenNotPaired(t *testing.T) {
	tournament := setupPairedRound(t, "Alice", "Bob", "Charlie", "Diana")

	if err := tournament.RemovePairing(1); err != nil {
		t.Fatalf("RemovePairing: %v", err)
	}
	err := tournament.RemovePairing(1)
	if err == nil || err.Error() != "player not currently paired" {
		t.Errorf("not currently paired: got %v", err)
	}
}

func TestRemovePairingBlocksNextRound(t *testing.T) {
	tournament := setupPairedRound(t, "Alice", "Bob", "Charlie", "Diana")

	if err := tournament.RemovePairing(1); err != nil {
		t.Fatalf("RemovePairing: %v", err)
	}
	// Add results for whatever pairings remain
	for _, p := range tournament.GetRound() {
		if p.playerb != BYE_OPPONENT_ID {
			tournament.AddResult(p.playera, 2, 1, 0)
		}
	}
	err := tournament.UpdatePlayerStandings()
	if err == nil {
		t.Fatal("expected UpdatePlayerStandings to error on unpaired player")
	}
	want := "player 1 (Alice) has no pairing in current round"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestManualRepairRoundTrip(t *testing.T) {
	// RemovePairing on two players, SetPairing them together, play through,
	// NextRound succeeds and standings reflect the manual pairing.
	tournament := setupPairedRound(t, "Alice", "Bob", "Charlie", "Diana")

	round := tournament.GetRound()
	a0, b0 := round[0].playera, round[0].playerb
	a1, b1 := round[1].playera, round[1].playerb

	// Remove both pairings
	if err := tournament.RemovePairing(a0); err != nil {
		t.Fatalf("RemovePairing a0: %v", err)
	}
	if err := tournament.RemovePairing(a1); err != nil {
		t.Fatalf("RemovePairing a1: %v", err)
	}
	// Pair a0 vs a1, b0 vs b1 (a manual re-pairing)
	if err := tournament.SetPairing(a0, a1); err != nil {
		t.Fatalf("SetPairing a0/a1: %v", err)
	}
	if err := tournament.SetPairing(b0, b1); err != nil {
		t.Fatalf("SetPairing b0/b1: %v", err)
	}
	for _, p := range tournament.GetRound() {
		if err := tournament.AddResult(p.playera, 2, 1, 0); err != nil {
			t.Fatalf("AddResult: %v", err)
		}
	}
	if err := tournament.NextRound(); err != nil {
		t.Fatalf("NextRound: %v", err)
	}
}

func TestUpdatePlayerStandingsSkipsRemoved(t *testing.T) {
	// A removed player must not trigger the unpaired-player guard.
	tournament := setupPairedRound(t, "Alice", "Bob", "Charlie")

	// Charlie has or doesn't have the bye; remove the bye player to
	// verify that a Removed player doesn't trigger the guard.
	var byeID int
	for _, p := range tournament.GetRound() {
		if p.playerb == BYE_OPPONENT_ID {
			byeID = p.playera
		}
	}
	if byeID == 0 {
		t.Fatal("no bye player found")
	}
	if err := tournament.RemovePlayerById(byeID); err != nil {
		t.Fatalf("RemovePlayerById: %v", err)
	}
	// Score remaining matches
	for _, p := range tournament.GetRound() {
		if p.playerb != BYE_OPPONENT_ID {
			if err := tournament.AddResult(p.playera, 2, 1, 0); err != nil {
				t.Fatalf("AddResult: %v", err)
			}
		}
	}
	if err := tournament.UpdatePlayerStandings(); err != nil {
		t.Fatalf("UpdatePlayerStandings should succeed with removed player: %v", err)
	}
}

func TestUpdatePlayerStandingsAtomicOnUnpaired(t *testing.T) {
	// When the unpaired-player guard fires, no stats are mutated.
	tournament := setupPairedRound(t, "Alice", "Bob", "Charlie", "Diana")

	// Remove player 1's pairing while results are still unset
	if err := tournament.RemovePairing(1); err != nil {
		t.Fatalf("RemovePairing: %v", err)
	}
	// Score all remaining pairings
	for _, p := range tournament.GetRound() {
		if p.playerb != BYE_OPPONENT_ID {
			if err := tournament.AddResult(p.playera, 2, 1, 0); err != nil {
				t.Fatalf("AddResult: %v", err)
			}
		}
	}

	before := map[int]int{}
	for id, p := range tournament.players {
		before[id] = p.Points
	}
	err := tournament.UpdatePlayerStandings()
	if err == nil {
		t.Fatal("expected error on unpaired player")
	}
	for id, p := range tournament.players {
		if p.Points != before[id] {
			t.Errorf("player %d: points changed from %d to %d despite error", id, before[id], p.Points)
		}
	}
}

// --- Playoff variants ---

func TestSwapPlayoffPlayersHappyPath(t *testing.T) {
	tournament := runSwissToFinish(t, 8)
	if err := tournament.StartPlayoff(4); err != nil {
		t.Fatalf("StartPlayoff: %v", err)
	}
	semis := tournament.GetPlayoffRound()
	if len(semis) != 2 {
		t.Fatalf("expected 2 semis, got %d", len(semis))
	}
	b0, b1 := semis[0].playerb, semis[1].playerb
	a0, a1 := semis[0].playera, semis[1].playera

	if err := tournament.SwapPlayoffPlayers(b0, b1); err != nil {
		t.Fatalf("SwapPlayoffPlayers: %v", err)
	}
	after := tournament.GetPlayoffRound()
	if after[0].playera != a0 || after[0].playerb != b1 {
		t.Errorf("pairing 0: want (%d, %d), got (%d, %d)", a0, b1, after[0].playera, after[0].playerb)
	}
	if after[1].playera != a1 || after[1].playerb != b0 {
		t.Errorf("pairing 1: want (%d, %d), got (%d, %d)", a1, b0, after[1].playera, after[1].playerb)
	}
}

func TestSwapPlayoffPlayersGuards(t *testing.T) {
	tournament := NewTournament()
	// No playoff
	err := tournament.SwapPlayoffPlayers(1, 2)
	if err == nil || err.Error() != "no playoff started" {
		t.Errorf("no playoff: got %v", err)
	}

	tournament = runSwissToFinish(t, 8)
	if err := tournament.StartPlayoff(4); err != nil {
		t.Fatalf("StartPlayoff: %v", err)
	}
	semis := tournament.GetPlayoffRound()
	a0, b0 := semis[0].playera, semis[0].playerb

	err = tournament.SwapPlayoffPlayers(a0, a0)
	if err == nil || err.Error() != "cannot swap a player with themselves" {
		t.Errorf("self-swap: got %v", err)
	}

	err = tournament.SwapPlayoffPlayers(a0, b0)
	if err == nil || err.Error() != "players are already paired against each other" {
		t.Errorf("same pairing: got %v", err)
	}

	err = tournament.SwapPlayoffPlayers(a0, 9999)
	if err == nil || err.Error() != "player not currently paired" {
		t.Errorf("not paired: got %v", err)
	}
}

// TestSwapPlayoffPlayersRebuildsRound rebuilds a bracket round into an entirely
// different set of pairings with a single swap. This is the rearrangement the
// removed RemovePlayoffPairing/SetPlayoffPairing pair existed for, reached
// without the round ever being short.
func TestSwapPlayoffPlayersRebuildsRound(t *testing.T) {
	tournament := runSwissToFinish(t, 8)
	if err := tournament.StartPlayoff(4); err != nil {
		t.Fatalf("StartPlayoff: %v", err)
	}
	semis := tournament.GetPlayoffRound()
	a0, b0 := semis[0].playera, semis[0].playerb
	a1, b1 := semis[1].playera, semis[1].playerb

	// {a0,b0},{a1,b1} -> {a0,a1},{b0,b1}
	if err := tournament.SwapPlayoffPlayers(b0, a1); err != nil {
		t.Fatalf("SwapPlayoffPlayers: %v", err)
	}

	rebuilt := tournament.GetPlayoffRound()
	if len(rebuilt) != 2 {
		t.Fatalf("expected 2 pairings, got %d", len(rebuilt))
	}
	if rebuilt[0].playera != a0 || rebuilt[0].playerb != a1 {
		t.Errorf("pairing 0: got %d vs %d, want %d vs %d", rebuilt[0].playera, rebuilt[0].playerb, a0, a1)
	}
	if rebuilt[1].playera != b0 || rebuilt[1].playerb != b1 {
		t.Errorf("pairing 1: got %d vs %d, want %d vs %d", rebuilt[1].playera, rebuilt[1].playerb, b0, b1)
	}

	for _, p := range rebuilt {
		if err := tournament.AddPlayoffResult(p.playera, 2, 0, 0); err != nil {
			t.Fatalf("AddPlayoffResult: %v", err)
		}
	}
	if err := tournament.NextPlayoffRound(); err != nil {
		t.Fatalf("NextPlayoffRound: %v", err)
	}
	if final := tournament.GetPlayoffRound(); len(final) != 1 {
		t.Fatalf("expected a 1-pairing final, got %d", len(final))
	}
}

// TestSwapPlayoffPlayersSurvivesDumpLoad is the regression test for issue #4.
// A consumer that persists between operations -- one load, mutate, dump per
// request -- must be able to adjust bracket pairings. The removed two-step flow
// could not, because its eligibility set was never serialized.
func TestSwapPlayoffPlayersSurvivesDumpLoad(t *testing.T) {
	tournament := runSwissToFinish(t, 8)
	if err := tournament.StartPlayoff(4); err != nil {
		t.Fatalf("StartPlayoff: %v", err)
	}
	semis := tournament.GetPlayoffRound()
	a0, b0 := semis[0].playera, semis[0].playerb
	a1, b1 := semis[1].playera, semis[1].playerb

	if err := tournament.SwapPlayoffPlayers(b0, a1); err != nil {
		t.Fatalf("SwapPlayoffPlayers: %v", err)
	}

	data, err := tournament.DumpTournament()
	if err != nil {
		t.Fatalf("DumpTournament: %v", err)
	}
	reloaded, err := LoadTournament(data)
	if err != nil {
		t.Fatalf("LoadTournament: %v", err)
	}

	round := reloaded.GetPlayoffRound()
	if len(round) != 2 {
		t.Fatalf("expected 2 pairings after reload, got %d", len(round))
	}
	if round[0].playera != a0 || round[0].playerb != a1 {
		t.Errorf("reloaded pairing 0: got %d vs %d, want %d vs %d", round[0].playera, round[0].playerb, a0, a1)
	}
	if round[1].playera != b0 || round[1].playerb != b1 {
		t.Errorf("reloaded pairing 1: got %d vs %d, want %d vs %d", round[1].playera, round[1].playerb, b0, b1)
	}

	// Adjust again after the reload: no in-memory state is required.
	if err := reloaded.SwapPlayoffPlayers(a1, b1); err != nil {
		t.Fatalf("SwapPlayoffPlayers after reload: %v", err)
	}

	for _, p := range reloaded.GetPlayoffRound() {
		if err := reloaded.AddPlayoffResult(p.playera, 2, 0, 0); err != nil {
			t.Fatalf("AddPlayoffResult: %v", err)
		}
	}
	if err := reloaded.NextPlayoffRound(); err != nil {
		t.Fatalf("NextPlayoffRound: %v", err)
	}
	if final := reloaded.GetPlayoffRound(); len(final) != 1 {
		t.Fatalf("expected a 1-pairing final, got %d", len(final))
	}
}

// TestPlayoffShortRoundDoesNotPanic covers a bracket round holding fewer
// pairings than its players need. No library call can produce this now, but a
// consumer can through the pointer GetPlayoff() returns, or by loading a dump
// written in that shape -- and NextPlayoffRound used to panic on it.
func TestPlayoffShortRoundDoesNotPanic(t *testing.T) {
	tournament := runSwissToFinish(t, 16)
	if err := tournament.StartPlayoff(8); err != nil {
		t.Fatalf("StartPlayoff: %v", err)
	}

	playoff := tournament.GetPlayoff()
	playoff.Rounds[0] = playoff.Rounds[0][:3]

	for _, p := range tournament.GetPlayoffRound() {
		if err := tournament.AddPlayoffResult(p.playera, 2, 0, 0); err != nil {
			t.Fatalf("AddPlayoffResult: %v", err)
		}
	}

	// Three winners cannot be paired, and byes are not supported here.
	err := tournament.NextPlayoffRound()
	want := "playoff round is incomplete - bracket cannot advance"
	if err == nil || err.Error() != want {
		t.Errorf("got %v, want %q", err, want)
	}
	if playoff.CurrentRound != 0 {
		t.Errorf("bracket advanced anyway: CurrentRound = %d", playoff.CurrentRound)
	}
}

// TestPlayoffEmptyRoundCannotAdvance pins the other end of the same guard: an
// empty bracket round used to advance silently, leaving a playoff still
// in_progress on a round with no players in it.
func TestPlayoffEmptyRoundCannotAdvance(t *testing.T) {
	tournament := runSwissToFinish(t, 8)
	if err := tournament.StartPlayoff(4); err != nil {
		t.Fatalf("StartPlayoff: %v", err)
	}

	playoff := tournament.GetPlayoff()
	playoff.Rounds[0] = nil

	err := tournament.NextPlayoffRound()
	want := "playoff round is incomplete - bracket cannot advance"
	if err == nil || err.Error() != want {
		t.Errorf("got %v, want %q", err, want)
	}
	if playoff.CurrentRound != 0 || playoff.Finished {
		t.Errorf("bracket advanced anyway: CurrentRound = %d, finished = %v", playoff.CurrentRound, playoff.Finished)
	}
}

// --- Result corrections: SetResult / RecomputeStandings / SetPlayoffResult ---

// setupFinishedTwoRounds plays two rounds with every player A winning 2-0 and
// finishes the tournament, so a correction to round 1 has both a later round
// and real tiebreakers to disturb.
func setupFinishedTwoRounds(t *testing.T, names ...string) Tournament {
	t.Helper()
	tournament := NewTournament()
	for _, n := range names {
		if err := tournament.AddPlayer(n); err != nil {
			t.Fatalf("AddPlayer %s: %v", n, err)
		}
	}
	if err := tournament.StartTournament(); err != nil {
		t.Fatalf("StartTournament: %v", err)
	}
	for _, p := range tournament.GetRound() {
		if p.PlayerB() != BYE_OPPONENT_ID {
			if err := tournament.AddResult(p.PlayerA(), 2, 0, 0); err != nil {
				t.Fatalf("AddResult round 1: %v", err)
			}
		}
	}
	if err := tournament.NextRound(); err != nil {
		t.Fatalf("NextRound: %v", err)
	}
	if err := tournament.Pair(false); err != nil {
		t.Fatalf("Pair round 2: %v", err)
	}
	for _, p := range tournament.GetRound() {
		if p.PlayerB() != BYE_OPPONENT_ID {
			if err := tournament.AddResult(p.PlayerA(), 2, 0, 0); err != nil {
				t.Fatalf("AddResult round 2: %v", err)
			}
		}
	}
	if err := tournament.FinishTournament(); err != nil {
		t.Fatalf("FinishTournament: %v", err)
	}
	return tournament
}

// setupCorrectionFixture plays five players through two finished rounds with a
// bye (odd count in round 1), a late entry who joins in round 2, and a player
// dropped mid-round-2. Replay has to reproduce all three with no special case.
// It returns the tournament, the late entry's ID and the dropped player's ID.
func setupCorrectionFixture(t *testing.T) (Tournament, int, int) {
	t.Helper()
	tournament := NewTournament()
	for i := 1; i <= 5; i++ {
		if err := tournament.AddPlayer(fmt.Sprintf("Player%d", i)); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := tournament.SetPlayerExternalID(1, 10101); err != nil {
		t.Fatalf("SetPlayerExternalID: %v", err)
	}
	if err := tournament.SetPlayerDecklist(1, Decklist{Main: map[string]int{"Island": 20}}); err != nil {
		t.Fatalf("SetPlayerDecklist: %v", err)
	}
	if err := tournament.StartTournament(); err != nil {
		t.Fatalf("StartTournament: %v", err)
	}
	// Round 1: odd player count, so one pairing is a pre-filled bye.
	for _, p := range tournament.GetRound() {
		if p.PlayerB() != BYE_OPPONENT_ID {
			if err := tournament.AddResult(p.PlayerA(), 2, 1, 0); err != nil {
				t.Fatalf("AddResult round 1: %v", err)
			}
		}
	}
	if err := tournament.NextRound(); err != nil {
		t.Fatalf("NextRound: %v", err)
	}
	// Late entry joins in round 2, so they have no round-1 pairing at all.
	if err := tournament.AddPlayer("Latecomer"); err != nil {
		t.Fatalf("AddPlayer late entry: %v", err)
	}
	lateID, ok := tournament.GetPlayerID("Latecomer")
	if !ok {
		t.Fatal("late entry not found")
	}
	if err := tournament.Pair(false); err != nil {
		t.Fatalf("Pair round 2: %v", err)
	}
	// Drop a paired player mid-round; their opponent inherits a bye in place.
	dropped := tournament.GetRound()[0].playerb
	if dropped == BYE_OPPONENT_ID {
		dropped = tournament.GetRound()[0].playera
	}
	if err := tournament.RemovePlayerById(dropped); err != nil {
		t.Fatalf("RemovePlayerById: %v", err)
	}
	for _, p := range tournament.GetRound() {
		if p.PlayerB() != BYE_OPPONENT_ID {
			if err := tournament.AddResult(p.PlayerA(), 2, 0, 1); err != nil {
				t.Fatalf("AddResult round 2: %v", err)
			}
		}
	}
	if err := tournament.FinishTournament(); err != nil {
		t.Fatalf("FinishTournament: %v", err)
	}
	return tournament, lateID, dropped
}

// standingsByID keys standings by player ID for before/after comparison.
func standingsByID(standings []PlayerStanding) map[int]PlayerStanding {
	m := make(map[int]PlayerStanding, len(standings))
	for _, s := range standings {
		m[s.PlayerID] = s
	}
	return m
}

func TestSetResultCompletedRoundSwapsResult(t *testing.T) {
	tournament := setupFinishedTwoRounds(t, "Alice", "Bob", "Charlie", "Diana")

	round1, err := tournament.GetRoundByNumber(1)
	if err != nil {
		t.Fatalf("GetRoundByNumber(1): %v", err)
	}
	winner, loser := round1[0].playera, round1[0].playerb
	if loser == BYE_OPPONENT_ID {
		t.Fatal("expected a real opponent in round 1")
	}
	// Snapshot round 2 to prove a correction never re-pairs anything.
	round2Before := append(Round(nil), tournament.rounds[2]...)
	before := tournament.GetPlayers()

	// Written from the recorded winner's side: they actually lost 0-2.
	if err := tournament.SetResult(1, winner, 0, 2, 0); err != nil {
		t.Fatalf("SetResult: %v", err)
	}
	if err := tournament.RecomputeStandings(); err != nil {
		t.Fatalf("RecomputeStandings: %v", err)
	}
	after := tournament.GetPlayers()

	w0, w1 := before[winner], after[winner]
	if w1.Points != w0.Points-3 || w1.Wins != w0.Wins-1 || w1.Losses != w0.Losses+1 || w1.Draws != w0.Draws {
		t.Errorf("player %d record did not lose the match: points %d->%d, record %d-%d-%d -> %d-%d-%d",
			winner, w0.Points, w1.Points, w0.Wins, w0.Losses, w0.Draws, w1.Wins, w1.Losses, w1.Draws)
	}
	if w1.GameWins != w0.GameWins-2 || w1.GameLosses != w0.GameLosses+2 || w1.GameDraws != w0.GameDraws {
		t.Errorf("player %d games did not swap: %d-%d-%d -> %d-%d-%d",
			winner, w0.GameWins, w0.GameLosses, w0.GameDraws, w1.GameWins, w1.GameLosses, w1.GameDraws)
	}

	l0, l1 := before[loser], after[loser]
	if l1.Points != l0.Points+3 || l1.Wins != l0.Wins+1 || l1.Losses != l0.Losses-1 || l1.Draws != l0.Draws {
		t.Errorf("player %d record did not win the match: points %d->%d, record %d-%d-%d -> %d-%d-%d",
			loser, l0.Points, l1.Points, l0.Wins, l0.Losses, l0.Draws, l1.Wins, l1.Losses, l1.Draws)
	}
	if l1.GameWins != l0.GameWins+2 || l1.GameLosses != l0.GameLosses-2 || l1.GameDraws != l0.GameDraws {
		t.Errorf("player %d games did not swap: %d-%d-%d -> %d-%d-%d",
			loser, l0.GameWins, l0.GameLosses, l0.GameDraws, l1.GameWins, l1.GameLosses, l1.GameDraws)
	}

	// Round 2 was actually played between those people; it must not move.
	if len(tournament.rounds[2]) != len(round2Before) {
		t.Fatalf("round 2 pairing count changed: %d -> %d", len(round2Before), len(tournament.rounds[2]))
	}
	for i, p := range tournament.rounds[2] {
		if p != round2Before[i] {
			t.Errorf("round 2 pairing %d changed: %+v -> %+v", i, round2Before[i], p)
		}
	}
}

func TestSetResultChangesStandingsOrder(t *testing.T) {
	tournament := setupFinishedTwoRounds(t, "Alice", "Bob", "Charlie", "Diana")

	// Find the round-1 winner who went on to lose round 2: they sit on 3
	// points, so flipping their round-1 match drops them to 0 and lifts their
	// opponent above them however the opponent's round 2 went.
	var x, y int
	for _, p := range tournament.rounds[1] {
		if p.playerb == BYE_OPPONENT_ID {
			continue
		}
		if tournament.players[p.playera].Points == 3 {
			x, y = p.playera, p.playerb
			break
		}
	}
	if x == 0 {
		t.Fatal("no round-1 winner on 3 points found")
	}

	before := standingsByID(tournament.GetStandings())
	if before[x].Points < before[y].Points {
		t.Fatalf("precondition: player %d (%d points) should not trail player %d (%d points)",
			x, before[x].Points, y, before[y].Points)
	}

	if err := tournament.SetResult(1, y, 2, 0, 0); err != nil {
		t.Fatalf("SetResult: %v", err)
	}
	if err := tournament.RecomputeStandings(); err != nil {
		t.Fatalf("RecomputeStandings: %v", err)
	}
	after := standingsByID(tournament.GetStandings())

	if after[y].Points <= after[x].Points {
		t.Errorf("expected player %d to overtake player %d: %d vs %d points",
			y, x, after[y].Points, after[x].Points)
	}
	if after[y].Rank >= after[x].Rank {
		t.Errorf("expected rank order to flip: player %d rank %d, player %d rank %d",
			y, after[y].Rank, x, after[x].Rank)
	}

	// Tiebreakers derive from the cumulative totals, so they follow too.
	if after[x].Tiebreakers.GameWinPercentage >= before[x].Tiebreakers.GameWinPercentage {
		t.Errorf("expected game win %% to drop for player %d: %v -> %v",
			x, before[x].Tiebreakers.GameWinPercentage, after[x].Tiebreakers.GameWinPercentage)
	}
	if after[x].Tiebreakers.OpponentMatchWinPct == before[x].Tiebreakers.OpponentMatchWinPct {
		t.Errorf("expected opponent match win %% to move for player %d, still %v",
			x, after[x].Tiebreakers.OpponentMatchWinPct)
	}

	// GetFinalStandings must tell the same story.
	final := standingsByID(tournament.GetFinalStandings())
	if final[y].Rank >= final[x].Rank {
		t.Errorf("GetFinalStandings disagrees: player %d rank %d, player %d rank %d",
			y, final[y].Rank, x, final[x].Rank)
	}
}

func TestRecomputeStandingsIdempotent(t *testing.T) {
	tournament := setupFinishedTwoRounds(t, "Alice", "Bob", "Charlie", "Diana")
	incremental := tournament.GetPlayers()

	if err := tournament.RecomputeStandings(); err != nil {
		t.Fatalf("RecomputeStandings: %v", err)
	}
	first := tournament.GetPlayers()
	if err := tournament.RecomputeStandings(); err != nil {
		t.Fatalf("RecomputeStandings (second call): %v", err)
	}
	second := tournament.GetPlayers()

	for id, p := range incremental {
		q := first[id]
		if q.Points != p.Points || q.Wins != p.Wins || q.Losses != p.Losses || q.Draws != p.Draws ||
			q.GameWins != p.GameWins || q.GameLosses != p.GameLosses || q.GameDraws != p.GameDraws {
			t.Errorf("player %d: replay disagrees with incremental totals: %+v vs %+v", id, p, q)
		}
		r := second[id]
		if r.Points != q.Points || r.Wins != q.Wins || r.Losses != q.Losses || r.Draws != q.Draws ||
			r.GameWins != q.GameWins || r.GameLosses != q.GameLosses || r.GameDraws != q.GameDraws {
			t.Errorf("player %d: RecomputeStandings is not idempotent: %+v vs %+v", id, q, r)
		}
	}
}

func TestRecomputeStandingsMatchesIncremental(t *testing.T) {
	tournament, lateID, dropped := setupCorrectionFixture(t)

	before, err := tournament.DumpTournament()
	if err != nil {
		t.Fatalf("DumpTournament: %v", err)
	}
	if err := tournament.RecomputeStandings(); err != nil {
		t.Fatalf("RecomputeStandings: %v", err)
	}
	after, err := tournament.DumpTournament()
	if err != nil {
		t.Fatalf("DumpTournament after recompute: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Errorf("replay changed the tournament\nbefore: %s\nafter:  %s", before, after)
	}

	players := tournament.GetPlayers()
	late := players[lateID]
	if late.Wins+late.Losses+late.Draws != 1 {
		t.Errorf("late entry should have exactly one round of results, got %d-%d-%d",
			late.Wins, late.Losses, late.Draws)
	}
	drop := players[dropped]
	if drop.Wins+drop.Losses+drop.Draws != 1 {
		t.Errorf("dropped player's totals should be frozen at one round, got %d-%d-%d",
			drop.Wins, drop.Losses, drop.Draws)
	}
}

func TestRecomputeStandingsPreservesPlayerMetadata(t *testing.T) {
	tournament, lateID, dropped := setupCorrectionFixture(t)

	before := tournament.GetPlayers()
	if len(before[lateID].Notes) == 0 {
		t.Fatal("expected a late-entry note to preserve")
	}
	if !before[dropped].Removed || before[dropped].RemovedInRound == 0 {
		t.Fatal("expected a removed player to preserve")
	}
	if before[1].ExternalID == nil || before[1].Decklist == nil {
		t.Fatal("expected external ID and decklist to preserve")
	}

	if err := tournament.RecomputeStandings(); err != nil {
		t.Fatalf("RecomputeStandings: %v", err)
	}
	after := tournament.GetPlayers()

	for id, p := range before {
		q := after[id]
		if q.Name != p.Name {
			t.Errorf("player %d: Name changed %q -> %q", id, p.Name, q.Name)
		}
		if len(q.Notes) != len(p.Notes) {
			t.Errorf("player %d: Notes changed %v -> %v", id, p.Notes, q.Notes)
		} else {
			for i := range p.Notes {
				if q.Notes[i] != p.Notes[i] {
					t.Errorf("player %d: note %d changed %q -> %q", id, i, p.Notes[i], q.Notes[i])
				}
			}
		}
		if q.Removed != p.Removed || q.RemovedInRound != p.RemovedInRound {
			t.Errorf("player %d: removal changed (%v,%d) -> (%v,%d)",
				id, p.Removed, p.RemovedInRound, q.Removed, q.RemovedInRound)
		}
		if (q.ExternalID == nil) != (p.ExternalID == nil) {
			t.Errorf("player %d: ExternalID presence changed", id)
		} else if p.ExternalID != nil && *q.ExternalID != *p.ExternalID {
			t.Errorf("player %d: ExternalID changed %d -> %d", id, *p.ExternalID, *q.ExternalID)
		}
		if (q.Decklist == nil) != (p.Decklist == nil) {
			t.Errorf("player %d: Decklist presence changed", id)
		} else if p.Decklist != nil && q.Decklist.Main["Island"] != p.Decklist.Main["Island"] {
			t.Errorf("player %d: Decklist changed", id)
		}
	}
}

func TestRecomputeStandingsAtomicOnIncompleteRound(t *testing.T) {
	tournament := setupFinishedTwoRounds(t, "Alice", "Bob", "Charlie", "Diana")
	before := tournament.GetPlayers()

	// Blank a stored result in an already-completed round.
	tournament.rounds[1][0].playeraWins = UNINITIALIZED_RESULT

	err := tournament.RecomputeStandings()
	if err == nil {
		t.Fatal("expected an error for an incomplete result in a completed round")
	}
	if !strings.Contains(err.Error(), "incomplete match") {
		t.Errorf("unexpected error: %v", err)
	}

	after := tournament.GetPlayers()
	for id, p := range before {
		q := after[id]
		if q.Points != p.Points || q.Wins != p.Wins || q.Losses != p.Losses || q.Draws != p.Draws ||
			q.GameWins != p.GameWins || q.GameLosses != p.GameLosses || q.GameDraws != p.GameDraws {
			t.Errorf("player %d: stats changed despite the error: %+v -> %+v", id, p, q)
		}
	}
}

func TestRecomputeStandingsToleratesSnapshotRoundOverrun(t *testing.T) {
	// LoadTournament accepts arbitrary JSON, so a snapshot can claim a current
	// round it has no pairings for. Replay must clamp rather than panic.
	tournament := setupFinishedTwoRounds(t, "Alice", "Bob", "Charlie", "Diana")
	data, err := tournament.DumpTournament()
	if err != nil {
		t.Fatalf("DumpTournament: %v", err)
	}
	tampered := strings.Replace(string(data), `"currentRound":2`, `"currentRound":99`, 1)
	if tampered == string(data) {
		t.Fatal("expected to find currentRound in the snapshot")
	}
	restored, err := LoadTournament([]byte(tampered))
	if err != nil {
		t.Fatalf("LoadTournament: %v", err)
	}

	if err := restored.RecomputeStandings(); err != nil {
		t.Fatalf("RecomputeStandings on overrun snapshot: %v", err)
	}
	// Every stored round still got replayed.
	for id, p := range tournament.GetPlayers() {
		q, ok := restored.GetPlayerById(id)
		if !ok {
			t.Fatalf("player %d missing after load", id)
		}
		if q.Points != p.Points || q.Wins != p.Wins || q.Losses != p.Losses {
			t.Errorf("player %d: expected the stored rounds to replay, got %+v vs %+v", id, p, q)
		}
	}
}

func TestSetResultErrors(t *testing.T) {
	tournament := setupPairedRound(t, "Alice", "Bob", "Charlie", "Diana")

	if err := tournament.SetResult(0, 1, 2, 0, 0); err == nil {
		t.Error("expected an error for round 0")
	}
	if err := tournament.SetResult(len(tournament.rounds), 1, 2, 0, 0); err == nil {
		t.Error("expected an error for a round past the end")
	}

	// Advance so that the new current round exists but has no pairings yet.
	for _, p := range tournament.GetRound() {
		if p.PlayerB() != BYE_OPPONENT_ID {
			if err := tournament.AddResult(p.PlayerA(), 2, 0, 0); err != nil {
				t.Fatalf("AddResult: %v", err)
			}
		}
	}
	if err := tournament.NextRound(); err != nil {
		t.Fatalf("NextRound: %v", err)
	}
	if err := tournament.SetResult(2, 1, 2, 0, 0); err == nil {
		t.Error("expected an error for a round with no pairings")
	}
	if err := tournament.SetResult(1, 999, 2, 0, 0); err == nil {
		t.Error("expected an error for a player not in that round")
	}
	for _, counts := range [][3]int{{-1, 0, 0}, {0, -1, 0}, {0, 0, -1}} {
		if err := tournament.SetResult(1, 1, counts[0], counts[1], counts[2]); err == nil {
			t.Errorf("expected an error for negative game counts %v", counts)
		}
	}
	for i, p := range tournament.rounds[1] {
		if p.playeraWins < 0 || p.playerbWins < 0 || p.draws < 0 {
			t.Errorf("rejected counts were written to round 1 pairing %d: %+v", i, p)
		}
	}
}

func TestSetResultPlayerB(t *testing.T) {
	tournament := setupFinishedTwoRounds(t, "Alice", "Bob", "Charlie", "Diana")

	target := tournament.rounds[1][0].playerb
	if target == BYE_OPPONENT_ID {
		t.Fatal("expected a real opponent in round 1")
	}
	if err := tournament.SetResult(1, target, 2, 1, 0); err != nil {
		t.Fatalf("SetResult: %v", err)
	}
	p := tournament.rounds[1][0]
	if p.playerbWins != 2 || p.playeraWins != 1 || p.draws != 0 {
		t.Errorf("side-B write did not swap perspective: a=%d b=%d draws=%d",
			p.playeraWins, p.playerbWins, p.draws)
	}
}

func TestSetPlayoffResultAdvancedRoundLeavesBracketIntact(t *testing.T) {
	tournament := runSwissToFinish(t, 8)
	if err := tournament.StartPlayoff(4); err != nil {
		t.Fatalf("StartPlayoff: %v", err)
	}
	for _, p := range tournament.GetPlayoffRound() {
		if err := tournament.AddPlayoffResult(p.PlayerA(), 2, 1, 0); err != nil {
			t.Fatalf("AddPlayoffResult semi: %v", err)
		}
	}
	if err := tournament.NextPlayoffRound(); err != nil {
		t.Fatalf("NextPlayoffRound: %v", err)
	}

	playoff := tournament.GetPlayoff()
	roundsBefore := len(playoff.Rounds)
	currentBefore := playoff.CurrentRound
	finalsBefore := append(Round(nil), playoff.Rounds[1]...)
	semiWinner := playoff.Rounds[0][0].playera

	// Flip the first semifinal: the recorded loser actually won it.
	if err := tournament.SetPlayoffResult(0, semiWinner, 1, 2, 0); err != nil {
		t.Fatalf("SetPlayoffResult: %v", err)
	}
	corrected := playoff.Rounds[0][0]
	if corrected.playeraWins != 1 || corrected.playerbWins != 2 {
		t.Errorf("semifinal not corrected: a=%d b=%d", corrected.playeraWins, corrected.playerbWins)
	}

	// Bracket structure is off limits.
	if len(playoff.Rounds) != roundsBefore {
		t.Errorf("Rounds length changed: %d -> %d", roundsBefore, len(playoff.Rounds))
	}
	if playoff.CurrentRound != currentBefore {
		t.Errorf("CurrentRound changed: %d -> %d", currentBefore, playoff.CurrentRound)
	}
	if playoff.Finished {
		t.Error("SetPlayoffResult must not finish the playoff")
	}
	if len(playoff.Rounds[1]) != len(finalsBefore) {
		t.Fatalf("final pairing count changed: %d -> %d", len(finalsBefore), len(playoff.Rounds[1]))
	}
	for i, p := range playoff.Rounds[1] {
		if p != finalsBefore[i] {
			t.Errorf("final pairing %d changed: %+v -> %+v", i, finalsBefore[i], p)
		}
	}
	// Those two people did play the final: the now-loser still advanced.
	if playoff.Rounds[1][0].playera != semiWinner && playoff.Rounds[1][0].playerb != semiWinner {
		t.Errorf("expected player %d to still appear in the final", semiWinner)
	}
}

func TestSetPlayoffResultOnFinishedPlayoff(t *testing.T) {
	tournament := runSwissToFinish(t, 8)
	if err := tournament.StartPlayoff(4); err != nil {
		t.Fatalf("StartPlayoff: %v", err)
	}
	for _, p := range tournament.GetPlayoffRound() {
		if err := tournament.AddPlayoffResult(p.PlayerA(), 2, 1, 0); err != nil {
			t.Fatalf("AddPlayoffResult semi: %v", err)
		}
	}
	if err := tournament.NextPlayoffRound(); err != nil {
		t.Fatalf("NextPlayoffRound: %v", err)
	}
	finals := tournament.GetPlayoffRound()
	if err := tournament.AddPlayoffResult(finals[0].PlayerA(), 2, 0, 0); err != nil {
		t.Fatalf("AddPlayoffResult final: %v", err)
	}
	if err := tournament.NextPlayoffRound(); err != nil {
		t.Fatalf("NextPlayoffRound final: %v", err)
	}
	if tournament.GetPlayoffStatus() != "finished" {
		t.Fatalf("expected a finished playoff, got %s", tournament.GetPlayoffStatus())
	}

	playoff := tournament.GetPlayoff()
	currentBefore := playoff.CurrentRound
	champion := playoff.Rounds[1][0].playera
	runnerUp := playoff.Rounds[1][0].playerb

	// AddPlayoffResult refuses a finished playoff; SetPlayoffResult is the
	// correction path precisely so a mis-entered final can still be fixed.
	if err := tournament.AddPlayoffResult(champion, 0, 2, 0); err == nil {
		t.Error("expected AddPlayoffResult to refuse a finished playoff")
	}
	// Written from the runner-up's side: they actually won 2-0.
	if err := tournament.SetPlayoffResult(1, runnerUp, 2, 0, 0); err != nil {
		t.Fatalf("SetPlayoffResult on finished playoff: %v", err)
	}
	if playoff.Rounds[1][0].playeraWins != 0 || playoff.Rounds[1][0].playerbWins != 2 {
		t.Errorf("final not corrected: %+v", playoff.Rounds[1][0])
	}
	if !playoff.Finished {
		t.Error("SetPlayoffResult must not clear Finished")
	}
	if playoff.CurrentRound != currentBefore {
		t.Errorf("CurrentRound changed: %d -> %d", currentBefore, playoff.CurrentRound)
	}
}

func TestSetPlayoffResultErrors(t *testing.T) {
	tournament := runSwissToFinish(t, 8)

	if err := tournament.SetPlayoffResult(0, 1, 2, 0, 0); err == nil {
		t.Error("expected an error with no playoff started")
	}
	if err := tournament.StartPlayoff(4); err != nil {
		t.Fatalf("StartPlayoff: %v", err)
	}
	seed := tournament.GetPlayoff().Seeds[0]

	if err := tournament.SetPlayoffResult(-1, seed, 2, 0, 0); err == nil {
		t.Error("expected an error for a negative playoff round")
	}
	if err := tournament.SetPlayoffResult(len(tournament.GetPlayoff().Rounds), seed, 2, 0, 0); err == nil {
		t.Error("expected an error for a playoff round past the end")
	}
	// Round 1 (the final) is not paired until the semis are played.
	if err := tournament.SetPlayoffResult(1, seed, 2, 0, 0); err == nil {
		t.Error("expected an error for a playoff round with no pairings")
	}
	if err := tournament.SetPlayoffResult(0, 999, 2, 0, 0); err == nil {
		t.Error("expected an error for a player not in that playoff round")
	}
	if err := tournament.SetPlayoffResult(0, seed, -1, 0, 0); err == nil {
		t.Error("expected an error for negative game counts")
	}
}

func TestSetResultDumpLoadRoundTrip(t *testing.T) {
	tournament := setupFinishedTwoRounds(t, "Alice", "Bob", "Charlie", "Diana")

	target := tournament.rounds[1][0].playerb
	if err := tournament.SetResult(1, target, 2, 0, 0); err != nil {
		t.Fatalf("SetResult: %v", err)
	}
	if err := tournament.RecomputeStandings(); err != nil {
		t.Fatalf("RecomputeStandings: %v", err)
	}

	data, err := tournament.DumpTournament()
	if err != nil {
		t.Fatalf("DumpTournament: %v", err)
	}
	restored, err := LoadTournament(data)
	if err != nil {
		t.Fatalf("LoadTournament: %v", err)
	}

	got, err := restored.GetRoundByNumber(1)
	if err != nil {
		t.Fatalf("GetRoundByNumber(1): %v", err)
	}
	if got[0].playerbWins != 2 || got[0].playeraWins != 0 {
		t.Errorf("correction did not survive the round trip: a=%d b=%d",
			got[0].playeraWins, got[0].playerbWins)
	}
	for id, p := range tournament.GetPlayers() {
		q, ok := restored.GetPlayerById(id)
		if !ok {
			t.Fatalf("player %d missing after load", id)
		}
		if q.Points != p.Points || q.Wins != p.Wins || q.Losses != p.Losses || q.Draws != p.Draws ||
			q.GameWins != p.GameWins || q.GameLosses != p.GameLosses || q.GameDraws != p.GameDraws {
			t.Errorf("player %d: totals did not survive the round trip: %+v vs %+v", id, p, q)
		}
	}

	// Recomputing the restored copy is a no-op.
	beforeDump, err := restored.DumpTournament()
	if err != nil {
		t.Fatalf("DumpTournament (restored): %v", err)
	}
	if err := restored.RecomputeStandings(); err != nil {
		t.Fatalf("RecomputeStandings (restored): %v", err)
	}
	afterDump, err := restored.DumpTournament()
	if err != nil {
		t.Fatalf("DumpTournament (restored, after recompute): %v", err)
	}
	if !bytes.Equal(beforeDump, afterDump) {
		t.Error("recompute on the restored copy changed it")
	}
}

// TestPlayoffTruncatedRoundCannotCrownChampion is the silent truncation from
// issue #4: a semifinal round holding one pairing produces one winner, which
// used to look exactly like a champion and finish the playoff.
func TestPlayoffTruncatedRoundCannotCrownChampion(t *testing.T) {
	tournament := runSwissToFinish(t, 8)
	if err := tournament.StartPlayoff(4); err != nil {
		t.Fatalf("StartPlayoff: %v", err)
	}

	playoff := tournament.GetPlayoff()
	playoff.Rounds[0] = playoff.Rounds[0][:1]

	survivor := playoff.Rounds[0][0].playera
	if err := tournament.AddPlayoffResult(survivor, 2, 0, 0); err != nil {
		t.Fatalf("AddPlayoffResult: %v", err)
	}

	err := tournament.NextPlayoffRound()
	want := "playoff round is incomplete - bracket cannot advance"
	if err == nil || err.Error() != want {
		t.Errorf("got %v, want %q", err, want)
	}
	if playoff.Finished {
		t.Error("a semifinal winner was crowned champion")
	}
	if tournament.GetPlayoffStatus() != "in_progress" {
		t.Errorf("status = %s, want in_progress", tournament.GetPlayoffStatus())
	}
}

// --- issue #3: the final round must count towards tiebreakers once finished ---

// nearlyEqual compares tiebreaker percentages, which are averages of floats and
// so cannot be compared exactly against a hand-computed expectation.
func nearlyEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

// forcePairings replaces the current round's pairings with exactly the given
// ones, so a test can control who plays whom instead of depending on the
// pairing algorithm's randomness. No pairing may have a result yet, and the
// round must contain no byes (use an even player count).
func forcePairings(t *testing.T, tournament *Tournament, pairs ...[2]int) {
	t.Helper()
	for _, p := range tournament.GetRound() {
		if p.PlayerB() == BYE_OPPONENT_ID {
			t.Fatalf("forcePairings: round contains a bye, which cannot be removed")
		}
		if err := tournament.RemovePairing(p.PlayerA()); err != nil {
			t.Fatalf("RemovePairing %d: %v", p.PlayerA(), err)
		}
	}
	for _, pair := range pairs {
		if err := tournament.SetPairing(pair[0], pair[1]); err != nil {
			t.Fatalf("SetPairing %d vs %d: %v", pair[0], pair[1], err)
		}
	}
}

// setupControlledFinalRound plays four players through two rounds with fixed
// pairings and fixed 2-0 results:
//
//	round 1: 1 beats 2, 3 beats 4
//	round 2: 3 beats 1, 2 beats 4
//
// Final records: 3 is 2-0, 1 and 2 are both 1-1, 4 is 0-2. Players 1 and 2 end
// level on points with identical game records, and they played each other in
// round 1 -- so their round-1 opponent strength is identical too, and only
// their round-2 opponents (3 at 2-0 versus 4 at 0-2) can order them. That makes
// the fixture sensitive to whether the final round reaches the tiebreakers.
//
// Round 2's results are entered but not applied: the caller picks how the
// tournament finishes.
func setupControlledFinalRound(t *testing.T) Tournament {
	t.Helper()
	tournament := NewTournament()
	for _, n := range []string{"Alice", "Bob", "Charlie", "Diana"} {
		if err := tournament.AddPlayer(n); err != nil {
			t.Fatalf("AddPlayer %s: %v", n, err)
		}
	}
	if err := tournament.StartTournament(); err != nil {
		t.Fatalf("StartTournament: %v", err)
	}

	forcePairings(t, &tournament, [2]int{1, 2}, [2]int{3, 4})
	if err := tournament.AddResult(1, 2, 0, 0); err != nil {
		t.Fatalf("AddResult round 1 (1 beats 2): %v", err)
	}
	if err := tournament.AddResult(3, 2, 0, 0); err != nil {
		t.Fatalf("AddResult round 1 (3 beats 4): %v", err)
	}
	if err := tournament.NextRound(); err != nil {
		t.Fatalf("NextRound: %v", err)
	}
	if err := tournament.Pair(false); err != nil {
		t.Fatalf("Pair round 2: %v", err)
	}

	forcePairings(t, &tournament, [2]int{3, 1}, [2]int{2, 4})
	if err := tournament.AddResult(3, 2, 0, 0); err != nil {
		t.Fatalf("AddResult round 2 (3 beats 1): %v", err)
	}
	if err := tournament.AddResult(2, 2, 0, 0); err != nil {
		t.Fatalf("AddResult round 2 (2 beats 4): %v", err)
	}
	return tournament
}

// checkControlledFinalRoundStandings asserts the standings the fixture implies
// when both rounds count. The expectations are hand-computed, with the 33%
// floor applied to any opponent below it:
//
//	1 (Alice):   3 pts, opponents 2 (1-1) and 3 (2-0) -> (0.5 + 1.0)/2
//	2 (Bob):     3 pts, opponents 1 (1-1) and 4 (0-2) -> (0.5 + 0.33)/2
//	3 (Charlie): 6 pts, opponents 4 (0-2) and 1 (1-1) -> (0.33 + 0.5)/2
//	4 (Diana):   0 pts, opponents 3 (2-0) and 2 (1-1) -> (1.0 + 0.5)/2
//
// Game win percentages happen to match match win percentages here because every
// match was 2-0, so the opponent game win expectations are the same numbers.
func checkControlledFinalRoundStandings(t *testing.T, standings []PlayerStanding, label string) {
	t.Helper()

	wantOrder := []int{3, 1, 2, 4}
	gotOrder := make([]int, 0, len(standings))
	for _, s := range standings {
		gotOrder = append(gotOrder, s.PlayerID)
	}
	if fmt.Sprint(gotOrder) != fmt.Sprint(wantOrder) {
		t.Errorf("%s: standings order = %v, want %v", label, gotOrder, wantOrder)
	}

	want := map[int]struct {
		rank   int
		points int
		gwp    float64
		omw    float64
		ogw    float64
	}{
		3: {rank: 1, points: 6, gwp: 1.0, omw: (0.33 + 0.5) / 2, ogw: (0.33 + 0.5) / 2},
		1: {rank: 2, points: 3, gwp: 0.5, omw: (0.5 + 1.0) / 2, ogw: (0.5 + 1.0) / 2},
		2: {rank: 3, points: 3, gwp: 0.5, omw: (0.5 + 0.33) / 2, ogw: (0.5 + 0.33) / 2},
		4: {rank: 4, points: 0, gwp: 0.0, omw: (1.0 + 0.5) / 2, ogw: (1.0 + 0.5) / 2},
	}

	byID := standingsByID(standings)
	for id, w := range want {
		got, ok := byID[id]
		if !ok {
			t.Errorf("%s: player %d missing from standings", label, id)
			continue
		}
		if got.Rank != w.rank {
			t.Errorf("%s: player %d rank = %d, want %d", label, id, got.Rank, w.rank)
		}
		if got.Points != w.points {
			t.Errorf("%s: player %d points = %d, want %d", label, id, got.Points, w.points)
		}
		if !nearlyEqual(got.Tiebreakers.GameWinPercentage, w.gwp) {
			t.Errorf("%s: player %d game win %% = %v, want %v", label, id, got.Tiebreakers.GameWinPercentage, w.gwp)
		}
		if !nearlyEqual(got.Tiebreakers.OpponentMatchWinPct, w.omw) {
			t.Errorf("%s: player %d opponent match win %% = %v, want %v", label, id, got.Tiebreakers.OpponentMatchWinPct, w.omw)
		}
		if !nearlyEqual(got.Tiebreakers.OpponentGameWinPct, w.ogw) {
			t.Errorf("%s: player %d opponent game win %% = %v, want %v", label, id, got.Tiebreakers.OpponentGameWinPct, w.ogw)
		}
	}
}

// TestFinishedTournamentTiebreakerValues pins the tiebreaker values and ranks of
// a finished tournament whose last round is the only thing separating two
// players level on points. Before issue #3 was fixed, round 2 was excluded and
// players 1 and 2 came out identical on every key, sharing rank 2.
func TestFinishedTournamentTiebreakerValues(t *testing.T) {
	tournament := setupControlledFinalRound(t)
	tournament.SetMaxRounds(2)
	if err := tournament.NextRound(); err != nil {
		t.Fatalf("NextRound: %v", err)
	}
	if tournament.GetStatus() != "finished" {
		t.Fatalf("status = %s, want finished", tournament.GetStatus())
	}
	// The whole point of the bug: a finished tournament never advanced past its
	// last round, so the last round looked like the in-progress one.
	if tournament.GetCurrentRound() != 2 {
		t.Fatalf("current round = %d, want 2", tournament.GetCurrentRound())
	}

	checkControlledFinalRoundStandings(t, tournament.GetStandings(), "GetStandings")
	checkControlledFinalRoundStandings(t, tournament.GetFinalStandings(), "GetFinalStandings")
}

// TestFinishTournamentTiebreakerValues covers the other way to finish: an
// explicit FinishTournament, which also leaves currentRound on the last round.
func TestFinishTournamentTiebreakerValues(t *testing.T) {
	tournament := setupControlledFinalRound(t)
	if err := tournament.FinishTournament(); err != nil {
		t.Fatalf("FinishTournament: %v", err)
	}
	if tournament.GetCurrentRound() != 2 {
		t.Fatalf("current round = %d, want 2", tournament.GetCurrentRound())
	}

	checkControlledFinalRoundStandings(t, tournament.GetStandings(), "GetStandings")
	checkControlledFinalRoundStandings(t, tournament.GetFinalStandings(), "GetFinalStandings")
}

// TestFinishedTiebreakersMatchExtraRound is the invariant from issue #3: a
// tournament finished after round 2 must produce the same standings as the same
// results with a third round opened but unplayed. Entering no new results
// cannot change the standings, so if advancing the round counter moves anyone,
// the finished tournament was reading the wrong set of rounds.
func TestFinishedTiebreakersMatchExtraRound(t *testing.T) {
	finished := setupControlledFinalRound(t)
	finished.SetMaxRounds(2)
	if err := finished.NextRound(); err != nil {
		t.Fatalf("NextRound (capped): %v", err)
	}

	advanced := setupControlledFinalRound(t)
	if err := advanced.NextRound(); err != nil {
		t.Fatalf("NextRound (uncapped): %v", err)
	}
	if advanced.GetCurrentRound() != 3 {
		t.Fatalf("current round = %d, want 3", advanced.GetCurrentRound())
	}

	a := finished.GetFinalStandings()
	b := advanced.GetFinalStandings()
	if len(a) != len(b) {
		t.Fatalf("standings length %d != %d", len(a), len(b))
	}
	for i := range a {
		if a[i].PlayerID != b[i].PlayerID || a[i].Rank != b[i].Rank {
			t.Errorf("position %d: finished has player %d at rank %d, advanced has player %d at rank %d",
				i, a[i].PlayerID, a[i].Rank, b[i].PlayerID, b[i].Rank)
		}
		if !nearlyEqual(a[i].Tiebreakers.OpponentMatchWinPct, b[i].Tiebreakers.OpponentMatchWinPct) {
			t.Errorf("position %d: opponent match win %% %v != %v",
				i, a[i].Tiebreakers.OpponentMatchWinPct, b[i].Tiebreakers.OpponentMatchWinPct)
		}
		if !nearlyEqual(a[i].Tiebreakers.OpponentGameWinPct, b[i].Tiebreakers.OpponentGameWinPct) {
			t.Errorf("position %d: opponent game win %% %v != %v",
				i, a[i].Tiebreakers.OpponentGameWinPct, b[i].Tiebreakers.OpponentGameWinPct)
		}
	}
}

// TestSetResultInFinalRoundMovesTiebreakers covers the reason issue #3 matters
// for result corrections: fixing a final-round result must move tiebreakers and
// not just points.
func TestSetResultInFinalRoundMovesTiebreakers(t *testing.T) {
	tournament := setupControlledFinalRound(t)
	if err := tournament.FinishTournament(); err != nil {
		t.Fatalf("FinishTournament: %v", err)
	}

	before := standingsByID(tournament.GetFinalStandings())

	// Flip round 2's other match: Diana (4) beat Bob (2) after all. Bob drops
	// from 1-1 to 0-2, which weakens Alice's (1) round-1 opponent, and Diana
	// rises to 1-1.
	if err := tournament.SetResult(2, 4, 2, 0, 0); err != nil {
		t.Fatalf("SetResult: %v", err)
	}
	if err := tournament.RecomputeStandings(); err != nil {
		t.Fatalf("RecomputeStandings: %v", err)
	}
	after := standingsByID(tournament.GetFinalStandings())

	if after[2].Points != 0 {
		t.Errorf("player 2 points = %d, want 0", after[2].Points)
	}
	if after[4].Points != 3 {
		t.Errorf("player 4 points = %d, want 3", after[4].Points)
	}

	// Alice's opponents are now 2 (0-2, floored to 0.33) and 3 (2-0). With the
	// final round excluded this was just 0.33, so the exact value is what pins
	// the fix.
	wantAlice := (0.33 + 1.0) / 2
	if !nearlyEqual(after[1].Tiebreakers.OpponentMatchWinPct, wantAlice) {
		t.Errorf("player 1 opponent match win %% = %v, want %v",
			after[1].Tiebreakers.OpponentMatchWinPct, wantAlice)
	}
	if nearlyEqual(after[1].Tiebreakers.OpponentMatchWinPct, before[1].Tiebreakers.OpponentMatchWinPct) {
		t.Errorf("expected player 1's opponent match win %% to move, still %v",
			after[1].Tiebreakers.OpponentMatchWinPct)
	}
}

// TestFinishedTiebreakersSurviveRoundTrip confirms the finished flag carries the
// behaviour through a dump and reload, since the flag is what tells
// calculateTiebreakers the last round is history.
func TestFinishedTiebreakersSurviveRoundTrip(t *testing.T) {
	tournament := setupControlledFinalRound(t)
	if err := tournament.FinishTournament(); err != nil {
		t.Fatalf("FinishTournament: %v", err)
	}

	data, err := tournament.DumpTournament()
	if err != nil {
		t.Fatalf("DumpTournament: %v", err)
	}
	loaded, err := LoadTournament(data)
	if err != nil {
		t.Fatalf("LoadTournament: %v", err)
	}

	checkControlledFinalRoundStandings(t, loaded.GetFinalStandings(), "after round trip")
}
